package main

import (
	"fmt"
	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/pcc-blackbox/models"
	"github.com/platinasystems/test"
	"testing"
	"time"
)

func updateNodes_installLLDP(t *testing.T) {
	t.Run("installLLDP", installLLDP)
}

// Install LLDP on all nodes
func installLLDP(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	if nodes, err := Pcc.GetNodeIds(); err == nil {
		installLLDPOnNodes(nodes)
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
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
	var roleId uint64
	if roleId, err = Pcc.FindRoleId(pcc.ROLE_LLDP, pcc.ROLE_DEFAULT); err == nil {
		err = setRolesToNodesAndCheck([]uint64{roleId}, pcc.ROLE_LLDPD, nodes, LLDP_TIMEOUT)
	}
	return
}
