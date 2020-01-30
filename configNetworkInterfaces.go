package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

var nodeIntfMap = make(map[uint64][]int64)

func configServerInterfaces(t *testing.T) {
	t.Run("configNetworkInterfaces", configNetworkInterfaces)
	t.Run("verifyNetworkInterfaces", verifyNetworkInterfaces)
	t.Run("verifyNetworkUp", verifyNetworkUp)
}

func configNetworkInterfaces(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var (
		err    error
		ifaces []*pcc.InterfaceDetail
	)
	for _, i := range Env.Invaders {
		id := NodebyHostIP[i.HostIp]
		ifaces, err = Pcc.GetIfacesByNodeId(id)
		if err != nil {
			assert.Fatalf("Error retrieving node %v id[%v] "+
				"interfaces", i.HostIp, NodebyHostIP[i.HostIp])
			return
		}
		var nodeIntfs []int64
		for _, intf := range ifaces {
			nodeIntfs = append(nodeIntfs, intf.Interface.Id)
		}
		configNodeInterfaces(t, i.HostIp, i.NetInterfaces, ifaces)
		nodeIntfMap[id] = nodeIntfs
	}
	for _, i := range Env.Servers {
		id := NodebyHostIP[i.HostIp]
		ifaces, err = Pcc.GetIfacesByNodeId(id)
		if err != nil {
			assert.Fatalf("Error retrieving node %v id[%v] "+
				"interfaces", i.HostIp, id)
			return
		}
		var nodeIntfs []int64
		for _, intf := range ifaces {
			nodeIntfs = append(nodeIntfs, intf.Interface.Id)
		}
		configNodeInterfaces(t, i.HostIp, i.NetInterfaces, ifaces)
		nodeIntfMap[id] = nodeIntfs
	}
}

func configNodeInterfaces(t *testing.T, HostIp string,
	serverInterfaces []netInterface, ifaces []*pcc.InterfaceDetail) {

	assert := test.Assert{t}
	var (
		iface        *pcc.InterfaceDetail
		ifaceRequest pcc.InterfaceRequest
		nodeId       uint64
		err          error
		ok           bool
	)

	if nodeId, ok = NodebyHostIP[HostIp]; !ok {
		assert.Fatalf("Failed to get nodeid for %v\n", HostIp)
		return
	}
	for j := 0; j < len(serverInterfaces); j++ {
		mac := serverInterfaces[j].MacAddr
		iface, err = Pcc.GetIfaceByMacAddress(mac, ifaces)
		if err != nil {
			assert.Fatalf("Error in retrieving interface having "+
				"MacAddress: %v for node %v id[%v]",
				mac, HostIp, nodeId)
			return
		}

		ifaceRequest.InterfaceId = iface.Interface.Id
		ifaceRequest.NodeId = nodeId
		ifaceRequest.Name = iface.Interface.Name
		ifaceRequest.Ipv4Addresses = serverInterfaces[j].Cidrs
		ifaceRequest.MacAddress = serverInterfaces[j].MacAddr
		ifaceRequest.ManagedByPcc = serverInterfaces[j].ManagedByPcc
		ifaceRequest.Gateway = serverInterfaces[j].Gateway
		ifaceRequest.Autoneg = serverInterfaces[j].Autoneg
		if ifaceRequest.Autoneg == "off" {
			ifaceRequest.Speed =
				json.Number(serverInterfaces[j].Speed)
		} else {
			ifaceRequest.Speed = ""
		}
		ifaceRequest.Mtu = json.Number(serverInterfaces[j].Mtu)
		ifaceRequest.AdminStatus = pcc.INTERFACE_STATUS_UP
		if serverInterfaces[j].IsManagement {
			ifaceRequest.IsManagement = "true"
		} else {
			ifaceRequest.IsManagement = "false"
		}
		fmt.Printf("Configuring node %v interface %v %v\n", nodeId,
			iface.Interface.Name, ifaceRequest)
		if err := Pcc.SetIface(ifaceRequest); err != nil {
			assert.Fatalf("Error setting interface %v for node "+
				"%v id[%v]: %v\n", ifaceRequest, HostIp,
				nodeId, err)
			return
		}
	}

	fmt.Printf("Apply interface changes for node %d\n", nodeId)
	if err = Pcc.ApplyIface(nodeId); err != nil {
		assert.Fatalf("Interface apply failed: %v\n", err)
		return
	}
}

func verifyNetworkInterfaces(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var nodesToCheck = make(map[uint64]uint64, len(nodeIntfMap))

	for id := range nodeIntfMap {
		nodesToCheck[id] = 1
	}

	timeout := time.After(5 * time.Minute)
	tick := time.Tick(5 * time.Second)
	for {
		select {
		case <-timeout:
			for id, _ := range nodesToCheck {
				intfs, found := nodeIntfMap[id]
				if !found {
					assert.Fatalf("map lookup failed %v\n",
						id)
					return
				}
				for _, i := range intfs {
					intf, _ := Pcc.GetIfaceById(id, i)
					state := intf.Interface.IntfState
					if state != pcc.Ready {
						fmt.Printf("failed to update"+
							" %v %v %v\n", id,
							intf.Interface.Name,
							state)
					}
				}
			}
			assert.Fatalf("time out updating interfaces\n")
			return
		case <-tick:
			for id, intfs := range nodeIntfMap {
				if _, found := nodesToCheck[id]; !found {
					continue
				}
				intf_count := len(intfs)
				intf_up := 0
				for _, i := range intfs {
					intf, err := Pcc.GetIfaceById(id, i)
					if err != nil {
						return
					}
					state := intf.Interface.IntfState
					switch state {
					case pcc.Ready:
						intf_up++
					case pcc.Queued:
					case pcc.Updating:
					case pcc.Unknown:
					default:
						assert.Fatalf("unexpected "+
							"IntfState %v\n",
							state)
						return
					}
				}
				if intf_count == intf_up {
					fmt.Printf("Node %v interfaces "+
						"updated\n", id)
					delete(nodesToCheck, id)
				}
			}
			if len(nodesToCheck) == 0 {
				return
			}
		}
	}
}

func verifyNetworkUp(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var nodesToCheck = make(map[uint64]uint64, len(nodeIntfMap))

	for id := range nodeIntfMap {
		nodesToCheck[id] = 1
	}

	timeout := time.After(5 * time.Minute)
	tick := time.Tick(5 * time.Second)
	for {
		select {
		case <-timeout:
			for id, _ := range nodesToCheck {
				intfs, found := nodeIntfMap[id]
				if !found {
					assert.Fatalf("map lookup failed %v\n",
						id)
					return
				}
				for _, i := range intfs {
					intf, err := Pcc.GetIfaceById(id, i)
					if err != nil {
						assert.Fatalf("getIfaceById: %v",
							err)
					}
					if !intf.Interface.ManagedByPcc {
						continue
					}
					ifName := intf.Interface.Name
					carrier := intf.Interface.CarrierStatus
					admin := intf.Interface.AdminStatus
					fmt.Printf("%v %v admin %v carrier "+
						"%v\n",
						id, ifName, admin, carrier)
				}
			}
			assert.Fatalf("time out updating interfaces\n")
			return
		case <-tick:
			for id, intfs := range nodeIntfMap {
				if _, found := nodesToCheck[id]; !found {
					continue
				}
				intf_count := len(intfs)
				intf_up := 0
				admin_down := 0
				for _, i := range intfs {
					intf, _ := Pcc.GetIfaceById(id, i)
					if !intf.Interface.ManagedByPcc {
						intf_count--
						continue
					}
					status := intf.Interface.AdminStatus
					if status == pcc.INTERFACE_STATUS_DOWN {
						admin_down++
					}
				}

				for _, i := range intfs {
					intf, _ := Pcc.GetIfaceById(id, i)
					if !intf.Interface.ManagedByPcc {
						continue
					}
					ifName := intf.Interface.Name
					carrier := intf.Interface.CarrierStatus
					admin := intf.Interface.AdminStatus
					if admin == pcc.INTERFACE_STATUS_DOWN {
						fmt.Printf("%v %v admin down\n",
							id, ifName)
					} else {
						fmt.Printf("%v %v carrier %v\n",
							id, ifName, carrier)
						if carrier == pcc.INTERFACE_STATUS_UP {
							intf_up++
						}
					}

				}
				if intf_up+admin_down == intf_count {
					fmt.Printf("Node %v interfaces "+
						"all UP\n", id)
					delete(nodesToCheck, id)
				}
			}
			if len(nodesToCheck) == 0 {
				return
			}
		}
	}
}
