package main

import (
	"encoding/json"
	"fmt"
	"github.com/platinasystems/test"
	"github.com/platinasystems/tiles/pccserver/models"
	"testing"
)

func updateNodes_installLLDP(t *testing.T) {
	t.Run("installLLDP", installLLDP)
}

func installLLDP(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var (
		body  []byte
		resp  HttpResp
		err   error
		check bool
	)
	for id := range Nodes {
		var (
			node models.NodeWithKubernetes
			data []byte
			lldp []uint64 = []uint64{2}
		)
		addReq := nodeAddReq{
			Host:  Nodes[id].Host,
			Id:    id,
			Roles: lldp,
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
			fmt.Printf("Update node %v failed\n%v\n", Nodes[id].Host, string(body))
			return
		}
		if err := json.Unmarshal(resp.Data, &node); err != nil {
			assert.Fatalf("%v\n%v\n", string(resp.Data), err)
			return
		}
	}

	//Check LLDP installation
	for id := range Nodes {
		fmt.Printf("Checking LLDP installation for nodeId:%v\n", id)
		check, err = checkLLDPInstallation(id)
		if err != nil {
			fmt.Printf("%v", err)
		}
		if check {
			fmt.Printf("LLDP correctly installed on nodeId:%v\n", id)
		}
	}
}
