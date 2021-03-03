package main

import (
	"github.com/platinasystems/pcc-blackbox/models"
	"github.com/platinasystems/test"
	"testing"
	"time"

	log "github.com/platinasystems/go-common/logs"
)

func addBrownfieldServers(t *testing.T) {
	t.Run("addBrownfieldServers", addServer)
}

func addServer(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now(), "addServer")
	CheckDependencies(t, res, Env.CheckServers)

	var nodes []node
	for i := range Env.Servers {
		nodes = append(nodes, Env.Servers[i].node)
	}

	log.AuctaLogger.Infof("adding %d servers\n", len(nodes))
	addNodesAndCheckStatus(t, nodes)
}
