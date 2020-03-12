package main

import (
	"fmt"
	"testing"
)

func addBrownfieldServers(t *testing.T) {
	t.Run("addBrownfieldServers", addServer)
}

func addServer(t *testing.T) {
	var nodes []node
	for i := range Env.Servers {
		nodes = append(nodes, Env.Servers[i].node)
	}

	fmt.Printf("adding %d servers\n", len(nodes))
	addNodesAndCheckStatus(t, nodes)
}
