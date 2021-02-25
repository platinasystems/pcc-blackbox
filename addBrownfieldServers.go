package main

import (
	"fmt"
	"testing"
)

func addBrownfieldServers(t *testing.T) {
	t.Run("addBrownfieldServers", addServer)
}

func addServer(t *testing.T) {
	var envNodes []node

	for i := range Env.Servers {
		envNodes = append(envNodes, Env.Servers[i].node)
	}

	fmt.Printf("adding %d servers\n", len(envNodes))
	nodesAdded := addEnvNodes(t, envNodes)
	checkAddNodesStatus(t, nodesAdded)
}
