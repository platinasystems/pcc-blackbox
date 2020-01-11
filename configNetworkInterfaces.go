package main

import (
	"encoding/json"
	"fmt"
	"github.com/platinasystems/test"
	"github.com/platinasystems/tiles/pccserver/models"
	"testing"
	"time"
)

func configServerInterfaces(t *testing.T) {
	t.Run("configNetworkInterfaces", configNetworkInterfaces)

}

func configNetworkInterfaces(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var (
		err    error
		ifaces []*models.InterfaceDetail
	)
	for _, i := range Env.Servers {
		ifaces, err = getIfacesByNodeId(NodebyHostIP[i.HostIp])
		if err != nil {
			assert.Fatalf("Error retrieving node %v id[%v] "+
				"interfaces", i.HostIp, NodebyHostIP[i.HostIp])
		}
		configNodeInterfaces(t, i.HostIp, i.NetInterfaces, ifaces)

	}
	for _, i := range Env.Invaders {
		ifaces, err = getIfacesByNodeId(NodebyHostIP[i.HostIp])
		if err != nil {
			assert.Fatalf("Error retrieving node %v id[%v] "+
				"interfaces", i.HostIp, NodebyHostIP[i.HostIp])
		}
		configNodeInterfaces(t, i.HostIp, i.NetInterfaces, ifaces)

	}
	time.Sleep(180 * time.Second) // lame
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
