package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/platinasystems/test"
	"github.com/platinasystems/tiles/pccserver/models"
)

var Nodes = make(map[uint64]*models.NodeWithKubernetes)

func addNodes(t *testing.T) {
	t.Run("addInvaders", addInvaders)
}

func addInvaders(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var (
		body []byte
		resp HttpResp
		err  error
	)
	for _, i := range Env.Invaders {
		var (
			node models.NodeWithKubernetes
			data []byte
		)
		addReq := nodeAddReq{
			Host:    i.HostIp,
			Managed: true,
		}
		endpoint := fmt.Sprintf("node/add")
		if data, err = json.Marshal(addReq); err != nil {
			assert.Fatalf("invalid struct for node add request")
		}
		if resp, body, err = pccGateway("POST", endpoint, data); err != nil {
			assert.Fatalf("%v\n%v\n", string(body), err)
			return
		}
		if resp.Status != 200 {
			assert.Fatalf("%v\n", string(body))
			fmt.Printf("add node %v failed\n%v\n", i.HostIp, string(body))
			return
		}
		if err := json.Unmarshal(resp.Data, &node); err != nil {
			assert.Fatalf("%v\n%v\n", string(resp.Data), err)
			return
		}
		if node.Id != 0 {
			Nodes[node.Id] = &node
			fmt.Printf("Add id %v to Nodes\n", node.Id)
		}
	}

	// wait for agent/collector to install
	time.Sleep(10 * time.Second)
	start := time.Now()
	done := false
	timeout := 90 * time.Second
	for !done {
		done = true
		for id, node := range Nodes {
			if node.NodeAvailabilityStatus != nil {
				if node.NodeAvailabilityStatus.ConnectionStatus == "online" {
					continue
				}
			}
			done = false
			endpoint := fmt.Sprintf("node/summary/%v", id)
			if resp, body, err = pccGateway("GET", endpoint, nil); err != nil {
				fmt.Printf("%v\n%v\n", string(body), err)
				continue
			}
			if resp.Status == 200 {
				if err := json.Unmarshal(resp.Data, &node); err == nil {
					name := fmt.Sprintf("nodeId:%v", id)
					if node.Name != "" {
						name = node.Name
					}
					fmt.Printf("%v is %v provisionStatus = %v \n",
						name, node.NodeAvailabilityStatus.ConnectionStatus, node.ProvisionStatus)
					done = node.NodeAvailabilityStatus.ConnectionStatus == "online"
					Nodes[id] = node
				}
			}
		}
		if !done {
			time.Sleep(10 * time.Second)
		}
		if time.Since(start) > timeout {
			break
		}
	}
	if !done {
		for _, node := range Nodes {
			if node.NodeAvailabilityStatus == nil {
				assert.Fatalf("node %v did not come online; provisionStatus = %v\n", node.Name, node.ProvisionStatus)
			} else {
				if node.NodeAvailabilityStatus.ConnectionStatus != "online" {
					assert.Fatalf("node %v did not come online; provisionStatus = %v\n", node.Name, node.ProvisionStatus)
				}
			}
		}
	}
}
