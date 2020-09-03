package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	pcc "github.com/platinasystems/pcc-blackbox/lib"
)

var outEnv testEnv

var nodes = make(map[uint64]*pcc.NodeDetailed)

func addTestTestNode(testNode *pcc.NodeDetailed) {
	var n node

	n.HostIp = testNode.Host
	n.BMCIp = testNode.Bmc
	n.BMCUser = testNode.BmcUser
	n.BMCUsers = []string{testNode.BmcUser}
	n.BMCPass = testNode.BmcPassword
	n.KeyAlias = []string{"test_0"}

	ifaces, err := Pcc.GetIfacesByNodeId(testNode.Id)
	if err != nil {
		fmt.Printf("error node %v: %v\n", testNode.Id, err)
		return
	}

	for _, intf := range ifaces {
		if intf.Interface.IsManagement || intf.Interface.ManagedByPcc {
			var net netInterface

			net.Name = intf.Interface.Name
			net.Gateway = intf.Interface.Gateway
			switch intf.Interface.Autoneg {
			case true:
				net.Autoneg = "true"
			case false:
				net.Autoneg = "false"
				net.Speed = intf.Interface.Speed
			}
			net.AdminStatus = intf.Interface.AdminStatus
			net.MacAddr = intf.Interface.MacAddress
			net.IsManagement = intf.Interface.IsManagement
			net.ManagedByPcc = intf.Interface.ManagedByPcc
			for _, addr := range intf.Interface.Ipv4Addresses {
				if strings.HasPrefix(addr, "203.0.113.") {
					// skip MaaS addresses
					continue
				}
				net.Cidrs = append(net.Cidrs, addr)
			}
			net.Mtu = strconv.Itoa(int(intf.Interface.Mtu))
			net.Fec = intf.Interface.FecType
			net.Media = intf.Interface.MediaType
			n.NetInterfaces = append(n.NetInterfaces, net)
		}
	}

	if strings.HasPrefix(testNode.Model, "PS-3001") {
		inv := invader{node: n}
		outEnv.Invaders = append(outEnv.Invaders, inv)
	} else {
		s := server{node: n}
		outEnv.Servers = append(outEnv.Servers, s)
	}

}

func addTestIpam() {
	subnets, err := Pcc.GetSubnetObj()
	if err != nil {
		fmt.Printf("Failed to GetSubnetObj: %v\n", err)
		return
	}
	if len(*subnets) == 0 {
		return
	}
	for _, ipam := range *subnets {
		var sub netIpam

		sub.Name = ipam.Name
		sub.Subnet = string(ipam.Subnet)
		sub.PubAccess = ipam.PubAccess
		sub.Routed = ipam.Routed
		outEnv.NetIpam = append(outEnv.NetIpam, sub)
	}
}

func addTestNetCluster() {
	netClusters, err := Pcc.GetNetCluster()
	if err != nil {
		fmt.Printf("Failed to GetSubnetObj: %v\n", err)
		return
	}
	if len(netClusters) == 0 {
		return
	}
	for _, net := range netClusters {
		var cluster netCluster
		var nodes []string

		cluster.Name = net.Name
		cluster.IgwPolicy = net.IgwPolicy
		controlCidr, _ := Pcc.GetSubnetObjId(net.ControlCIDRId)
		cluster.ControlCIDR = controlCidr.Name
		dataCidr, _ := Pcc.GetSubnetObjId(net.DataCIDRId)
		cluster.DataCIDR = dataCidr.Name
		for _, node := range net.Nodes {
			nodes = append(nodes, node.MgmtIPv4Address)
		}
		cluster.Nodes = nodes
		outEnv.NetCluster = append(outEnv.NetCluster, cluster)
	}
}

func genEnv() {

	outEnv.PccIp = Env.PccIp

	nodes, err := Pcc.GetNodesDetail()
	if err != nil {
		fmt.Printf("Failed to GetNodes: %v\n", err)
		return
	}
	for _, testNode := range nodes {
		addTestTestNode(testNode)
	}

	addTestIpam()
	addTestNetCluster()

	data, err := json.MarshalIndent(outEnv, "", "    ")
	if err == nil {
		fmt.Printf("\n%v\n", string(data))
	} else {
		fmt.Printf("Error marshal to json: %v\n", err)
	}
}
