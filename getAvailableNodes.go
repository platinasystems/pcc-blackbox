package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/platinasystems/pcc-blackbox/models"

	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

func getNodes(t *testing.T) {
	t.Run("getAvailableNodes", getAvailableNodes)
	t.Run("TestNodeGroups", testNodeGroups)
}

func getAvailableNodes(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	assert := test.Assert{t}
	if nodes, err := Pcc.GetNodes(); err == nil {
		for i := 0; i < len(*nodes); i++ {
			node := (*nodes)[i]
			id := node.Id
			Nodes[id] = &node
			NodebyHostIP[node.Host] = id
			Env.setNodeId(node.Host, id)
		}
	} else {
		msg := fmt.Sprintf("Error getting nodes: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
}

func testNodeGroups(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	assert := test.Assert{t}

	var req pcc.NodeGroupRequest

	req.Name = "my-test-group"
	req.Description = "just a silly test"

	err := Pcc.AddNodeGroup(&req)
	if err != nil {
		msg := fmt.Sprintf("Error getting nodes: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
	id := req.Id

	group, err := Pcc.GetNodeGroupId(id)
	if err != nil {
		msg := fmt.Sprintf("Error GetNodeGroupId: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	newDescription := "change it"

	group.Description = newDescription
	err = Pcc.UpdateNodeGroup(&group)
	if err != nil {
		msg := fmt.Sprintf("Error UpdateNodeGroupId: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	groups, err := Pcc.GetNodeGroups()
	if err != nil {
		msg := fmt.Sprintf("Error UpdateNodeGroupId: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Errorf("Error GetNodeGroups: %v", err)
		assert.FailNow()
	}
	found := false
	for _, g := range groups {
		if g.ID == id {
			if g.Description == newDescription {
				found = true
				break
			}
		}
	}
	if !found {
		msg := "Error finding updated group"
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	err = Pcc.DeleteNodeGroup(group.ID)
	if err != nil {
		msg := fmt.Sprintf("Error DeleteNodeGroup: %v", err)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
}
