package main

import (
	"testing"

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
		assert.Fatalf("Error getting nodes: %v\n", err)
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
		assert.Fatalf("Error AddNodeGroup: %v\n", err)
		return
	}
	id := req.Id

	group, err := Pcc.GetNodeGroupId(id)
	if err != nil {
		assert.Fatalf("Error GetNodeGroupId: %v\n", err)
		return
	}

	newDescription := "change it"

	group.Description = newDescription
	err = Pcc.UpdateNodeGroup(&group)
	if err != nil {
		assert.Fatalf("Error UpdateNodeGroupId: %v\n", err)
		return
	}

	groups, err := Pcc.GetNodeGroups()
	if err != nil {
		assert.Fatalf("Error GetNodeGroups: %v\n", err)
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
		assert.Fatalf("Error finding updated group\n")
		return
	}

	err = Pcc.DeleteNodeGroup(group.ID)
	if err != nil {
		assert.Fatalf("Error DeleteNodeGroup: %v\n", err)
		return
	}
}
