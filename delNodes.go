package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/platinasystems/test"
)

func delAllNodes(t *testing.T) {
	t.Run("delNodes", delNodes)
	t.Run("validateDeleteNodes", validateDeleteNodes)
}

func delNodes(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var err error

	for id, _ := range Nodes {
		err = Pcc.DelNode(id)
		if err != nil {
			assert.Fatalf("Failed to delete %v: %v\n", id, err)
			return
		}
	}

}

func validateDeleteNodes(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	var err error

	// wait for node to be removed
	time.Sleep(5 * time.Second)
	start := time.Now()
	done := false
	timeout := 300 * time.Second
	for !done {
		done = true
		for id, node := range Nodes {
			done = false
			err = Pcc.GetNodeSummary(id, node)
			if err != nil {
				if err.Error() == "no such node" {
					fmt.Printf("%v deleted\n", node.Name)
					delete(Nodes, id)
					if len(Nodes) == 0 {
						done = true
					}
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
			assert.Fatalf("node %v provisionStatus = %v was not "+
				"deleted\n", node.Name, node.ProvisionStatus)
		}
	}
}
