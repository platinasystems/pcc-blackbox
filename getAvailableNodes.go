package main

import (
	"testing"

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
		// log.AuctaLogger.Errorf("Error getting nodes: %v\n", err)
		log.AuctaLogger.Errorf("Error getting nodes: %v\n", err)
		assert.FailNow()
	}
}

func testNodeGroups(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	var req pcc.NodeGroupRequest

	req.Name = "my-test-group"
	req.Description = "just a silly test"

	err := Pcc.AddNodeGroup(&req)
	if err != nil {
		log.AuctaLogger.Errorf("Error AddNodeGroup: %v\n", err)
		assert.FailNow()
		return
	}
	id := req.Id

	group, err := Pcc.GetNodeGroupId(id)
	if err != nil {
		log.AuctaLogger.Errorf("Error GetNodeGroupId: %v\n", err)
		assert.FailNow()
		return
	}

	newDescription := "change it"

	group.Description = newDescription
	err = Pcc.UpdateNodeGroup(&group)
	if err != nil {
		log.AuctaLogger.Errorf("Error UpdateNodeGroupId: %v\n", err)
		assert.FailNow()
		return
	}

	groups, err := Pcc.GetNodeGroups()
	if err != nil {
		log.AuctaLogger.Errorf("Error GetNodeGroups: %v\n", err)
		assert.FailNow()
		return
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
		log.AuctaLogger.Errorf("Error finding updated group\n")
		assert.FailNow()
		return
	}

	err = Pcc.DeleteNodeGroup(group.ID)
	if err != nil {
		log.AuctaLogger.Errorf("Error DeleteNodeGroup: %v\n", err)
		assert.FailNow()
		return
	}
}
