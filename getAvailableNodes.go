package main

import (
	"testing"

	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

func getNodes(t *testing.T) {
	t.Run("getAvailableNodes", getAvailableNodes)
}

func getAvailableNodes(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var (
		err   error
		nodes []*pcc.NodeWithKubernetes
	)

	nodes, err = Pcc.GetNodesWithKubernetes()
	if err != nil {
		assert.Fatalf("Error geting nodes: %v\n", err)
		return
	}

	for i := 0; i < len(nodes); i++ {
		Nodes[nodes[i].Id] = nodes[i]
		NodebyHostIP[nodes[i].Host] = nodes[i].Id
	}
}
