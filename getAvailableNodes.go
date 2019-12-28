package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/platinasystems/test"
	"github.com/platinasystems/tiles/pccserver/models"
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
		NodebyHostIP[nodes[i].Host] = nodes[i].Id
	}
}

func getNodeSummary(id uint64, node *models.NodeWithKubernetes) (err error) {
	var (
		resp HttpResp
	)
	endpoint := fmt.Sprintf("pccserver/node/summary/%v", id)
	if resp, _, err = pccGateway("GET", endpoint, nil); err != nil {
		return
	}
	if resp.Status == 200 {
		err = json.Unmarshal(resp.Data, node)
		if err != nil {
			return
		}
		return
	}
	err = fmt.Errorf("%v", resp.Message)
	return
}

func getProvisionStatus(id uint64) (status string, err error) {
	var (
		resp HttpResp
	)
	endpoint := fmt.Sprintf("pccserver/node/%v/provisionStatus", id)
	if resp, _, err = pccGateway("GET", endpoint, nil); err != nil {
		return
	}
	if resp.Status == 200 {
		status = string(resp.Data) // status has double quotes
		return
	}
	err = fmt.Errorf("%v", resp.Message)
	return
}
