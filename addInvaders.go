package main

import (
	"fmt"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
	"strings"
	"sync"
	"testing"
	"time"
)

func addClusterHeads(t *testing.T) {
	t.Run("addInvaders", addInvaders)
}

func addInvaders(t *testing.T) {
	var envNodes []node

	for i := range Env.Invaders {
		envNodes = append(envNodes, Env.Invaders[i].node)
	}

	fmt.Printf("adding %d invaders\n", len(envNodes))
	nodesAdded := addEnvNodes(t, envNodes)
	checkAddNodesStatus(t, nodesAdded)
}

func addEnvNodes(t *testing.T, envNodes []node) (nodesAdded []uint64) {
	var (
		newNode *pcc.NodeDetailed
		err     error
	)

	for _, n := range envNodes {
		if NodebyHostIP[n.HostIp] != 0 {
			if Nodes[NodebyHostIP[n.HostIp]] != nil {
				fmt.Printf("the node %s was already added\n",
					n.HostIp)
				continue
			}
		}
		newNode = new(pcc.NodeDetailed)
		newNode.Host = n.HostIp
		newNode.Managed = new(bool)
		*newNode.Managed = true

		fmt.Printf("adding the node %s\n", newNode.Host)
		if err = Pcc.AddNode(newNode); err != nil {
			err = fmt.Errorf("add node %s failed: %v",
				newNode.Host, err)
			t.Fatal(err)
			return
		}
		Nodes[newNode.Id] = newNode
		NodebyHostIP[n.HostIp] = newNode.Id
		nodesAdded = append(nodesAdded, newNode.Id)
		fmt.Printf("Add id %d to Nodes. Mapping hostIP %v to id %d\n",
			newNode.Id, newNode.Host, newNode.Id)
	}
	return
}

func checkAddNodesStatus(t *testing.T, nodeIdsAdded []uint64) {
	var err error
	test.SkipIfDryRun(t)

	if len(nodeIdsAdded) < 1 {
		return
	}

	//Check Agent and collector installation function.
	// FIXME add a channel for stopping on error
	waitInstallation := func(timeout time.Duration, app string, nodeId uint64, from *time.Time) {
		fmt.Printf("Checking %s installation for nodeId:%v from %s \n",
			app, nodeId, from.String())
		check, waitErr := Pcc.WaitForInstallation(nodeId, timeout,
			app, "", from)
		if waitErr != nil {
			fmt.Printf("\n%v\n", waitErr)
			err = waitErr
		}
		if check {
			fmt.Printf("%s correctly installed on nodeId:%v\n",
				app, nodeId)
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(nodeIdsAdded))

	// add the node and wait for the services.
	// FIXME add a channel for stopping on error
	addNodeWaitDone := func(id uint64, nodeNumbers int) {
		defer wg.Done()

		eventsFrom := time.Now()
		waitInstallation(LLDP_TIMEOUT*time.Duration(nodeNumbers),
			LLDP_NOTIFICATION, id, &eventsFrom)
		waitInstallation(DEFAULT_TIMEOUT*time.Duration(nodeNumbers),
			SELF_HEAL_NOTIFICATION, id, &eventsFrom)
		start := time.Now()
		timeout := time.Duration(300*len(nodeIdsAdded)) * time.Second
		var (
			connection         string
			previousConnection string
			ignore             error
			status             string
		)
		host := Nodes[id].Host
		for true {
			time.Sleep(10 * time.Second)
			status, ignore = Pcc.GetProvisionStatus(id)
			if ignore != nil { // early check for add fail
				if strings.Contains(status, "Add node failed") {
					err = fmt.Errorf("%s for node %d",
						status, id)
					t.Fatal(err)
					return
				}
			}
			connection, ignore = Pcc.GetNodeConnectionStatus(id)
			if ignore == nil {
				switch connection { // FIXME use models
				case "online":
					fmt.Printf("the node %d:%s is %v\n",
						id, host, connection)
					return
				case "", "NoRunningService":
					// wait for the next cycle
				default:
					err = fmt.Errorf("Unable to add node %s",
						connection)
					t.Fatal(err)
					return
				}
				if previousConnection != "" &&
					previousConnection != connection {
					fmt.Printf("node %d:%s connection "+
						"status switched from %s to %s\n",
						id, host, previousConnection,
						connection)
				}
				previousConnection = connection
			} else {
				fmt.Printf("error getting the connection "+
					"status for node %d:%s %v\n",
					id, host, err)
			}

			if time.Since(start) > timeout {
				err = fmt.Errorf("timeout for node addition %d",
					id)
				t.Fatal(err)
			}

			if err != nil {
				return
			}
		}
	}

	for i := range nodeIdsAdded { // check nodes in parallel
		go addNodeWaitDone(nodeIdsAdded[i], len(nodeIdsAdded))
	}

	wg.Wait() // wait for all addition

	if err != nil {
		fmt.Println(fmt.Sprintf("Error adding nodes %v", err))
		t.Fatal(err)
	}

}
