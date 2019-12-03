package main

import (
	"encoding/json"
	"fmt"
	"github.com/platinasystems/test"
	"github.com/platinasystems/tiles/pccserver/models"
	"testing"
)

func getNodes(t *testing.T) {
	t.Run("getAvailableNodes", getAvailableNodes)
}

func getAvailableNodes(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var (
		body  []byte
		resp  HttpResp
		err   error
		nodes []models.NodeWithKubernetes
	)
	endpoint := fmt.Sprintf("pccserver/node")
	if resp, body, err = pccGateway("GET", endpoint, nil); err != nil {
		assert.Fatalf("%v\n%v\n", string(body), err)
		return
	}
	if resp.Status != 200 {
		assert.Fatalf("%v\n", string(body))
		fmt.Printf("Get node list failed\n%v\n", string(body))
		return
	}
	if err := json.Unmarshal(resp.Data, &nodes); err != nil {
		assert.Fatalf("%v\n%v\n", string(resp.Data), err)
		return
	}
	for i := 0; i < len(nodes); i++ {
		Nodes[nodes[i].Id] = &nodes[i]
		fmt.Printf("Update Nodes[%v]\n", nodes[i].Id)
		NodebyHostIP[nodes[i].Host] = nodes[i].Id
		fmt.Printf("Mapping hostIP %v to id %v\n", nodes[i].Host, nodes[i].Id)
	}
}
