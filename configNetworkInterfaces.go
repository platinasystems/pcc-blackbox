package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/platinasystems/test"
	"github.com/platinasystems/tiles/pccserver/models"
)

var nodeIntfMap = make(map[uint64][]int64)

func configServerInterfaces(t *testing.T) {
	t.Run("configNetworkInterfaces", configNetworkInterfaces)
	t.Run("verifyNetworkInvertaces", verifyNetworkInterfaces)
}

func configNetworkInterfaces(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var (
		err    error
		ifaces []*models.InterfaceDetail
	)
	for _, i := range Env.Servers {
		id := NodebyHostIP[i.HostIp]
		ifaces, err = getIfacesByNodeId(id)
		nodeIntfs := make([]int64, len(ifaces))
		for j, intf := range ifaces {
			nodeIntfs[j] = intf.Interface.Id
		}
		if err != nil {
			assert.Fatalf("Error retrieving node %v id[%v] "+
				"interfaces", i.HostIp, id)
		}
		configNodeInterfaces(t, i.HostIp, i.NetInterfaces, ifaces)
		nodeIntfMap[id] = nodeIntfs

	}
	for _, i := range Env.Invaders {
		id := NodebyHostIP[i.HostIp]
		ifaces, err = getIfacesByNodeId(id)
		nodeIntfs := make([]int64, len(ifaces))
		for j, intf := range ifaces {
			nodeIntfs[j] = intf.Interface.Id
		}
		if err != nil {
			assert.Fatalf("Error retrieving node %v id[%v] "+
				"interfaces", i.HostIp, NodebyHostIP[i.HostIp])
		}
		configNodeInterfaces(t, i.HostIp, i.NetInterfaces, ifaces)
		nodeIntfMap[id] = nodeIntfs
	}
}

func configNodeInterfaces(t *testing.T, HostIp string,
	serverInterfaces []netInterface, ifaces []*models.InterfaceDetail) {

	assert := test.Assert{t}
	var (
		iface        *models.InterfaceDetail
		ifaceRequest models.InterfaceRequest
		err          error
	)

	for j := 0; j < len(serverInterfaces); j++ {
		fmt.Printf("Looking for %v MacAddress\n",
			serverInterfaces[j].MacAddr)
		iface, err = getIfaceByMacAddress(serverInterfaces[j].MacAddr,
			ifaces)
		if err != nil {
			assert.Fatalf("Error in retrieving interface having "+
				"MacAddress: %v for node %v id[%v]",
				serverInterfaces[j].MacAddr, HostIp,
				NodebyHostIP[HostIp])
		}
		if iface == nil {
			continue
		}
		ifaceRequest = models.InterfaceRequest{
			InterfaceId:   iface.Interface.Id,
			NodeId:        NodebyHostIP[HostIp],
			Name:          iface.Interface.Name,
			Ipv4Addresses: serverInterfaces[j].Cidrs,
			MacAddress:    serverInterfaces[j].MacAddr,
			ManagedByPcc:  serverInterfaces[j].ManagedByPcc,
			Gateway:       serverInterfaces[j].Gateway,
			Autoneg:       serverInterfaces[j].Autoneg,
			Speed:         json.Number(serverInterfaces[j].Speed),
			Mtu:           json.Number(serverInterfaces[j].Mtu),
		}
		if iface.Interface.IsManagement {
			ifaceRequest.IsManagement = "true"
		}
		if setIface(ifaceRequest) == nil {
			continue
		}
		assert.Fatalf("Error setting interface %v for node %v id[%v]",
			ifaceRequest, HostIp, NodebyHostIP[HostIp])
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
				}
				for _, i := range intfs {
					intf, _ := getIfaceById(id, i)
					state := intf.Interface.IntfState
					if state != "ready" {
						fmt.Printf("failed to update"+
							" %v %v %v\n", id,
							intf.Interface.Name,
							state)
					}
				}
			}
			assert.Fatalf("time out updating interfaces\n")
		case <-tick:
			for id, intfs := range nodeIntfMap {
				if _, found := nodesToCheck[id]; !found {
					continue
				}
				notDone := false
				for _, i := range intfs {
					intf, err := getIfaceById(id, i)
					if err != nil {
						return
					}
					state := intf.Interface.IntfState
					switch state {
					case "ready":
					case "queued":
						notDone = true
					case "updating":
						notDone = true
					case "offline":
					default:
						assert.Fatalf("unexpected "+
							"IntfState %v\n",
							state)
					}
				}
				if notDone == false {
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
