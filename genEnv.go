package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/platinasystems/tiles/pccserver/models"
)

var outEnv testEnv

func addTestTestNode(testNode *models.NodeWithKubernetes) {
	var n node

	n.HostIp = testNode.Host
	n.BMCIp = testNode.Bmc
	n.BMCUser = testNode.BmcUser
	n.BMCUsers = []string{testNode.BmcUser}
	n.BMCPass = testNode.BmcPassword
	n.KeyAlias = []string{"test_0"}

	ifaces, err := getIfacesByNodeId(testNode.Id)
	if err != nil {
		fmt.Printf("error node %v: %v\n", testNode.Id, err)
		return
	}

	for _, intf := range ifaces {
		if intf.Interface.IsManagement || intf.Interface.ManagedByPcc {
			var net netInterface

			net.Name = intf.Interface.Name
			net.Gateway = intf.Interface.Gateway
			if intf.Interface.Autoneg {
				net.Autoneg = "on"
			} else {
				net.Autoneg = "off"
				net.Speed = intf.Interface.Speed
			}
			net.MacAddr = intf.Interface.MacAddress
			net.IsManagement = intf.Interface.IsManagement
			net.ManagedByPcc = intf.Interface.ManagedByPcc
			net.Cidrs = intf.Interface.Ipv4Addresses
			net.Mtu = strconv.Itoa(int(intf.Interface.Mtu))
			n.NetInterfaces = append(n.NetInterfaces, net)
		}
	}

	if testNode.Model == "PS-3001-32C-AFA" {
		inv := invader{node: n}
		outEnv.Invaders = append(outEnv.Invaders, inv)
	} else {
		s := server{node: n}
		outEnv.Servers = append(outEnv.Servers, s)
	}
}

func genEnv() {

	outEnv.PccIp = Env.PccIp

	for _, testNode := range Nodes {
		addTestTestNode(testNode)
	}

	data, err := json.MarshalIndent(outEnv, "", "    ")
	if err == nil {
		fmt.Printf("\n%v\n", string(data))
	} else {
		fmt.Printf("Error marshal to json: %v\n", err)
	}
}
