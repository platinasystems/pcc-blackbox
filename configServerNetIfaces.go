package main

import (
	"encoding/json"
	"fmt"
	"github.com/platinasystems/test"
	"github.com/platinasystems/tiles/pccserver/models"
	"testing"
)

func configServerIntefaces(t *testing.T) {
	t.Run("configNetworkIntefaces", configNetworkIntefaces)

}

func configNetworkIntefaces(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var (
		body []byte
		resp HttpResp
		err  error
	)
	for _, i := range Env.Servers {
		var (
			node models.NodeWithKubernetes
			data []byte
		)
		addReq := nodeAddReq{
			Host:        i.HostIp,
			Id:          NodebyHostIP[i.HostIp],
			Bmc:         i.BMCIp,
			BmcUser:     i.BMCUser,
			BmcUsers:    i.BMCUsers,
			BmcPassword: i.BMCPass,
			AdminUser:   "admin",
			SSHKeys:     []uint64{3}, //todo: get key for key manager
			Managed:     true,
		}
		endpoint := fmt.Sprintf("pccserver/node/update")
		if data, err = json.Marshal(addReq); err != nil {
			assert.Fatalf("invalid struct for node update request")
		}
		if resp, body, err = pccGateway("PUT", endpoint, data); err != nil {
			assert.Fatalf("%v\n%v\n", string(body), err)
			return
		}
		if resp.Status != 200 {
			assert.Fatalf("%v\n", string(body))
			fmt.Printf("Update node %v failed\n%v\n", i.HostIp, string(body))
			return
		}
		if err := json.Unmarshal(resp.Data, &node); err != nil {
			assert.Fatalf("%v\n%v\n", string(resp.Data), err)
			return
		}
	}

}
