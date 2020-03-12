package main

import (
	"github.com/platinasystems/test"
	"testing"
)

func updateNodes_installLLDP(t *testing.T) {
	t.Run("installLLDP", installLLDP)
}

// Install LLDP on all nodes
func installLLDP(t *testing.T) {
	test.SkipIfDryRun(t)
	if nodes, err := Pcc.GetNodeIds(); err == nil {
		installLLDPOnNodes(nodes)
	} else {
		panic(err)
	}
}

// Install LLDP on all invaders
func installLLDPOnInvaders(t *testing.T) {
	if nodes, err := Pcc.GetInvaderIds(); err == nil {
		if err = installLLDPOnNodes(nodes); err != nil {
			t.Fatal(err)
		}
	} else {
		t.Fatal(err)
	}
}

// Install LLDP on nodes
func installLLDPOnNodes(nodes []uint64) (err error) {
	err = setRolesToNodesAndCheck([]uint64{2}, LLDP_NOTIFICATION, nodes, LLDP_TIMEOUT)
	return
}
