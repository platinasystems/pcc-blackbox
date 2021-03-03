package main

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/platinasystems/pcc-blackbox/models"

	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

func addClusterHeads(t *testing.T) {
	t.Run("addInvaders", addInvaders)
}

func addInvaders(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now(), "addInvaders")
	CheckDependencies(t, res, Env.CheckInvaders)

	var nodes []node

	for i := range Env.Invaders {
		nodes = append(nodes, Env.Invaders[i].node)
	}

	log.AuctaLogger.Infof("adding %d invaders\n", len(nodes))
	addNodesAndCheckStatus(t, nodes)
}

// add nodes fom list
func addNodesAndCheckStatus(t *testing.T, nodes []node) {
	var err error
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now(), "addNodesAndCheckStatus")

	//Check Agent and collector installation function. FIXME add a channel for stopping on error
	waitInstallation := func(timeout time.Duration, app string, nodeId uint64, from *time.Time) {
		log.AuctaLogger.Infof("Checking %s installation for nodeId:%v from %s \n", app, nodeId, from.String())
		check, waitErr := Pcc.WaitForInstallation(nodeId, timeout, app, "", from)
		if waitErr != nil {
			log.AuctaLogger.Errorf("\n%v\n", waitErr)
			err = waitErr
		}
		if check {
			log.AuctaLogger.Infof("%s correctly installed on nodeId:%v\n", app, nodeId)
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(nodes))

	addNodeAndWait := func(n node, nodeNumbers int) { // add the node and wait for the services. FIXME add a channel for stopping on error
		defer wg.Done()
		var (
			node         pcc.NodeDetailed
			routineError error
		)
		node.Host = n.HostIp
		node.Managed = new(bool)
		*node.Managed = true

		if Nodes[NodebyHostIP[node.Host]] == nil { // add the node
			log.AuctaLogger.Infof("adding the node %s\n", node.Host)
			if routineError = Pcc.AddNode(&node); routineError == nil {
				n.Id = node.Id
				Nodes[node.Id] = &node
				NodebyHostIP[node.Host] = node.Id
				log.AuctaLogger.Infof("Add id %d to Nodes. Mapping hostIP %v to id %d\n", node.Id, node.Host, node.Id)

				eventsFrom := time.Now()
				waitInstallation(LLDP_TIMEOUT*time.Duration(nodeNumbers), LLDP_NOTIFICATION, node.Id, &eventsFrom)
				waitInstallation(DEFAULT_TIMEOUT*time.Duration(nodeNumbers), SELF_HEAL_NOTIFICATION, node.Id, &eventsFrom)
				start := time.Now()
				timeout := time.Duration(300*len(nodes)) * time.Second
				var (
					connection         string
					previousConnection string
					ignore             error
					status             string
				)
				for true {
					time.Sleep(10 * time.Second)
					if status, ignore = Pcc.GetProvisionStatus(node.Id); ignore == nil { // early check for add fail
						if strings.Contains(status, "Add node failed") {
							msg := fmt.Sprintf("%s for node %d", status, node.Id)
							res.SetTestFailure(msg)
							log.AuctaLogger.Error(msg)
							t.FailNow()
						}
					}
					if connection, ignore = Pcc.GetNodeConnectionStatus(node.Id); ignore == nil {
						switch connection { // FIXME use models
						case "online":
							fmt.Printf("the node %d:%s is online\n", node.Id, node.Host)
							return
						case "", "NoRunningService": // wait for the next cycle
						default:
							msg := fmt.Sprintf("Unable to add the node %s", connection)
							res.SetTestFailure(msg)
							log.AuctaLogger.Error(msg)
							t.FailNow()

						}
						if previousConnection != "" && previousConnection != connection {
							log.AuctaLogger.Infof("the node %d:%s connection status switched from %s to %s\n", node.Id, node.Host, previousConnection, connection)
						}

						previousConnection = connection
					} else {
						log.AuctaLogger.Errorf("error getting the connection status for node %d:%s %v\n", node.Id, node.Host, err)
					}

					if time.Since(start) > timeout {
						msg := fmt.Sprintf("timeout for node addition %d", node.Id)
						res.SetTestFailure(msg)
						log.AuctaLogger.Error(msg)
						t.FailNow()
					}

					if err != nil {
						return
					}
				}
			} else {
				msg := fmt.Sprintf("add node %s failed\n%v\n", node.Host, routineError)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				t.FailNow()
			}
		} else {
			log.AuctaLogger.Warnf("the node %s was already added\n", node.Host)
		}
	}

	for i := range nodes { // add nodes in parallel
		go addNodeAndWait(nodes[i], len(nodes))
	}

	wg.Wait() // wait for all addition

	if err != nil {
		msg := fmt.Sprintf("Error adding nodes %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}
}
