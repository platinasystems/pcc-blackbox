package main

import (
	"encoding/json"
	"fmt"
	"github.com/platinasystems/test"
	"github.com/platinasystems/tiles/pccserver/models"
	"testing"
	"time"
)

func configServerIntefaces(t *testing.T) {
	t.Run("configNetworkIntefaces", configNetworkIntefaces)

}

func configNetworkIntefaces(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var (
		err          error
		ifaces       []*models.InterfaceDetail
		iface        *models.InterfaceDetail
		ifaceRequest models.InterfaceRequest
	)
	for _, i := range Env.Servers {
		ifaces, err = getIfacesByNodeId(NodebyHostIP[i.HostIp])
		if err != nil {
			assert.Fatalf("Error in retrieving node %v id[%v] interfaces", i.HostIp, NodebyHostIP[i.HostIp])
		}
		for j := 0; j < len(i.NetInterfaces); j++ {
			fmt.Printf("Looking for %v MacAddress\n", i.NetInterfaces[j].MacAddr)
			iface, err = getIfaceByMacAddress(i.NetInterfaces[j].MacAddr, ifaces)
			if err != nil {
				assert.Fatalf("Error in retrieving interface having MacAddress: %v for node %v id[%v]", i.NetInterfaces[j].MacAddr, i.HostIp, NodebyHostIP[i.HostIp])
			}
			if iface != nil {
				ifaceRequest = models.InterfaceRequest{
					InterfaceId:   iface.Interface.Id,
					NodeId:        NodebyHostIP[i.HostIp],
					Name:          iface.Interface.Name,
					Ipv4Addresses: i.NetInterfaces[j].Cidrs,
					MacAddress:    i.NetInterfaces[j].MacAddr,
					ManagedByPcc:  i.NetInterfaces[j].ManagedByPcc,
					Gateway:       i.NetInterfaces[j].Gateway,
					Autoneg:       i.NetInterfaces[j].Autoneg,
					Speed:         json.Number(i.NetInterfaces[j].Speed),
					Mtu:           json.Number(i.NetInterfaces[j].Mtu),
				}
				if iface.Interface.IsManagement {
					ifaceRequest.IsManagement = "true"
				}
				if setIface(ifaceRequest) != nil {
					assert.Fatalf("Error in setting interface %v for node %v id[%v]", ifaceRequest, i.HostIp, NodebyHostIP[i.HostIp])
				}
			}
		}
	}
	time.Sleep(180 * time.Second)
}
