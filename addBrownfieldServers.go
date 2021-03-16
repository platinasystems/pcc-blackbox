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
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res, Env.CheckServers)

	var envNodes []node
	for i := range Env.Servers {
		envNodes = append(envNodes, Env.Servers[i].node)
	}

	log.AuctaLogger.Infof("adding %d servers\n", len(envNodes))
	nodesAdded := addEnvNodes(t, envNodes)
	checkAddNodesStatus(t, nodesAdded)
}
