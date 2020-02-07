package main

import (
	"fmt"
	"testing"
	"time"

	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

func updateNodes_installLLDP(t *testing.T) {
	t.Run("installLLDP", installLLDP)
}

func installLLDP(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var (
		err   error
		check bool
	)

	var isLLDPInNodes = make(map[uint64]bool)
	for id := range Nodes {
		var (
			addReq pcc.NodeWithKubernetes
			lldp   []uint64 = []uint64{2}
		)

		addReq.Host = Nodes[id].Host
		addReq.Id = id
		addReq.RoleIds = lldp

		isLLDPInNodes[id] = Pcc.IsAppInstalled(id, "lldpd")
		if !isLLDPInNodes[id] {
			_, err = Pcc.UpdateNode(addReq)
			if err != nil {
				assert.Fatalf("Failed to install LLDP on id "+
					"%v : %v", id, err)
				return
			}
		} else {
			fmt.Printf("LLDP already installed in nodeId:%v\n", id)
		}
	}

	from := time.Now()
	//Check LLDP installation
	for id := range Nodes {
		if !isLLDPInNodes[id] {
			fmt.Printf("Checking LLDP installation for nodeId:"+
				"%v\n", id)

			check, err = checkGenericInstallation(id, LLDP_TIMEOUT,
				LLDP_NOTIFICATION, from)
			if err != nil {
				assert.Fatalf("Failed checking LLDP on %v "+
					": %v", id, err)
				return
			}
			if check {
				fmt.Printf("LLDP correctly installed on "+
					"nodeId:%v\n", id)
			}
		}
	}
}
