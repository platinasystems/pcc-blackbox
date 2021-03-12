package main

import (
	"fmt"
	log "github.com/platinasystems/go-common/logs"
	"github.com/platinasystems/pcc-blackbox/models"
	"github.com/platinasystems/test"
	"testing"
	"time"
)

func delAllNodes(t *testing.T) {
	t.Run("delNodes", delNodes)
	t.Run("validateDeleteNodes", validateDeleteNodes)
}

func delNodes(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	if _, err := Pcc.DeleteNodes(true); err != nil {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}
}

func validateDeleteNodes(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	assert := test.Assert{t}

	// wait for node to be removed
	time.Sleep(5 * time.Second)
	start := time.Now()
	timeout := 60 * 8 * time.Second // FIXME move to config
	for true {
		if nodes, err := Pcc.GetNodes(); err == nil {
		l1:
			for k := range Nodes { // clear the map or remove deleted nodes
				for _, node := range *nodes {
					if node.Id == k {
						continue l1
					}
				}
				delete(Nodes, k)
			}
		} else {
			log.AuctaLogger.Errorf("error getting nodes %v\n", err)
		}
		if len(Nodes) == 0 {
			log.AuctaLogger.Infof("all nodes have been deleted")
			return
		} else if len(Nodes) > 0 {
			time.Sleep(5 * time.Second)
		}
		if time.Since(start) > timeout {
			log.AuctaLogger.Infof("delAllNodes timeout\n")
			for _, node := range Nodes {
				msg := fmt.Sprintf("node %v provisionStatus = %v was not "+
					"deleted\n", node.Name, node.ProvisionStatus)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				assert.FailNow()
			}

			return
		}
	}
}
