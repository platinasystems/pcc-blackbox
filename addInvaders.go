package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

func addClusterHeads(t *testing.T) {
	t.Run("addInvaders", addInvaders)
}

func addInvaders(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var (
		err   error
		check bool
	)
	nodesToCheck := make([]uint64, 0)
	for _, i := range Env.Invaders {
		if Nodes[NodebyHostIP[i.HostIp]] != nil {
			continue
		}
		var node pcc.NodeWithKubernetes

		node, err := Pcc.AddNode(i.HostIp, true)
		if err != nil {
			assert.Fatalf("Error adding node %v: %v\n", i.HostIp,
				err)
			return
		}
		if node.Id != 0 {
			node.Invader = true
			nodesToCheck = append(nodesToCheck, node.Id)
			Nodes[node.Id] = &node
			fmt.Printf("Add Cluster Head id %v to Nodes\n", node.Id)
			NodebyHostIP[node.Host] = node.Id
			fmt.Printf("Mapping hostIP %v to id %v\n",
				node.Host, node.Id)
		}
	}

	// early check for add fail
	time.Sleep(10 * time.Second)
	for id := range Nodes {
		if status, err := Pcc.GetProvisionStatus(id); err == nil {
			if strings.Contains(status, "Add node failed") {
				assert.Fatalf("%v for %v\n", status, id)
			}
		}
	}

	from := time.Now()
	//Check Agent installation
	//SERIAL - to be improved
	for i := 0; i < len(nodesToCheck); i++ {
		check = false
		fmt.Printf("Checking Agent installation for nodeId:%v\n",
			nodesToCheck[i])
		check, err = checkGenericInstallation(nodesToCheck[i],
			AGENT_TIMEOUT, AGENT_NOTIFICATION, from)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
		if check {
			fmt.Printf("AGENT correctly installed on nodeId:%v\n",
				nodesToCheck[i])
		}
	}

	from = time.Now()
	//Check Collector installation
	for i := 0; i < len(nodesToCheck); i++ {
		fmt.Printf("Checking Collector installation for nodeId:%v\n",
			nodesToCheck[i])
		check, err = checkGenericInstallation(nodesToCheck[i],
			COLLECTOR_TIMEOUT, COLLECTOR_NOTIFICATION, from)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
		if check {
			fmt.Printf("COLLECTOR correctly installed on nodeId:"+
				":%v\n", nodesToCheck[i])
		}
	}

	// waiting for node becomes online
	start := time.Now()
	done := false
	timeout := 180 * time.Second
	for !done {
		done = true
		for id, node := range Nodes {
			var connection string
			status := node.NodeAvailabilityStatus
			if status != nil {
				connection, _ = Pcc.GetNodeConnectionStatus(node)
				if connection == "online" {
					continue
				}
			}
			done = false
			err = Pcc.GetNodeSummary(id, node)
			if err != nil {
				fmt.Printf("node %v, error: %v\n", id, err)
				continue
			}
			name := fmt.Sprintf("node:%v", id)
			if node.Name != "" {
				name = node.Name
			}
			fmt.Printf("%v is %v provisionStatus = %v \n", name,
				connection, node.ProvisionStatus)
			if node.ProvisionStatus == "Add node failed" {
				assert.Fatalf("%v for %v\n",
					node.ProvisionStatus, name)
			}
			done = connection == "online"
			Nodes[id] = node
		}

		if !done {
			time.Sleep(10 * time.Second)
		}
		if time.Since(start) > timeout {
			assert.Fatalf("timeout")
			break
		}
	}

	if !done {
		for _, node := range Nodes {
			if node.NodeAvailabilityStatus == nil {
				assert.Fatalf("node %v did not come online; "+
					"provisionStatus = %v\n",
					node.Name, node.ProvisionStatus)
			} else {
				connection, _ := Pcc.GetNodeConnectionStatus(node)
				if connection != "online" {
					assert.Fatalf("node %v did not come "+
						"online; provisionStatus = "+
						"%v\n", node.Name,
						node.ProvisionStatus)
				}
			}
		}
	}
}
