package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/platinasystems/test"
)

func delNodes(t *testing.T) {
	t.Run("delAllNodes", delAllNodes)
}

func delAllNodes(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var (
		body []byte
		resp HttpResp
		err  error
	)
	for id, _ := range Nodes {
		endpoint := fmt.Sprintf("pccserver/node/%v", id)
		if resp, body, err = pccGateway("DELETE", endpoint, nil); err != nil {
			assert.Fatalf("%v\n%v\n", string(body), err)
			return
		}
		if resp.Status != 200 {
			assert.Fatalf("%v\n", string(body))
			fmt.Printf("delete node %v failed\n%v\n", id, string(body))
			return
		}
	}

	// wait for node to be removed
	time.Sleep(5 * time.Second)
	start := time.Now()
	done := false
	timeout := 300 * time.Second
	for !done {
		done = true
		for id, node := range Nodes {
			done = false
			endpoint := fmt.Sprintf("pccserver/node/summary/%v", id)
			if resp, body, err = pccGateway("GET", endpoint, nil); err != nil {
				fmt.Printf("%v\n%v\n", string(body), err)
				continue
			}
			if resp.Status == 200 {
				if err := json.Unmarshal(resp.Data, &node); err == nil {
					fmt.Printf("%v provisionStatus = %v\n", node.Name, node.ProvisionStatus)
					Nodes[id] = node
				}
			}
			if resp.Status == 400 {
				fmt.Printf("%v deleted\n", node.Name)
				delete(Nodes, id)
				if len(Nodes) == 0 {
					done = true
				}
			}
		}
		if !done {
			time.Sleep(5 * time.Second)
		}
		if time.Since(start) > timeout {
			fmt.Printf("delAllNodes timeout\n")
			break
		}
	}
	if !done {
		for _, node := range Nodes {
			assert.Fatalf("node %v provisionStatus = %v was not deleted\n", node.Name, node.ProvisionStatus)
		}
	}
}
