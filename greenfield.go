package main

import (
	"fmt"
	"os/exec"
	"sync"
	"testing"
	"time"

	log "github.com/platinasystems/go-common/logs"
	"github.com/platinasystems/pcc-blackbox/models"
)

// Delete servers and pxeboot (in parallel)
func addGreenfieldServers(t *testing.T) {

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	log.AuctaLogger.Info("GREENFIELD: executing the pxeboot for the servers")
	var (
		wg        sync.WaitGroup
		mainError error
	)

	if err := Pcc.DeleteServers(true); err != nil { // wait for the deletion
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

	invaders, err := Pcc.GetInvaders()
	if err != nil {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
		return
	}
	for _, i := range *invaders {
		log.AuctaLogger.Infof("Update MaaS for node [%v]", i.Id)
		err = Pcc.UpdateMaas(&i)
		if err != nil {
			msg := fmt.Sprintf("%v", err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
			return
		}
	}

	wg.Add(len(Env.Servers))

	pxeboot := func(server node) {
		defer wg.Done()
		bmc := server.BMCIp
		host := server.HostIp
		log.AuctaLogger.Infof(fmt.Sprintf("executing the pxeboot for server [%s]", bmc))
		cmd := exec.Command("ipmitool", "-I", "lanplus", "-H", bmc, "-U", "ADMIN", "-P", "ADMIN", "chassis", "bootdev", "pxe")
		if err := cmd.Run(); err == nil {
			cmd = exec.Command("ipmitool", "-I", "lanplus", "-H", bmc, "-U", "ADMIN", "-P", "ADMIN", "chassis", "power", "cycle")
			if err = cmd.Run(); err == nil {
				for i := 1; i <= 30; i++ { //wait for the node
					time.Sleep(time.Second * 15)
					if nodes, err := Pcc.GetNodes(); err == nil {
						for _, node := range *nodes {
							if node.Bmc == bmc {
								log.AuctaLogger.Infof(fmt.Sprintf("the pxeboot for the server [%s]ql node added with id [%d]", bmc, node.Id))
								node.Host = host
								server.Id = node.Id
								if err = Pcc.UpdateNode(&node); err != nil { // Set the host address for the greenfield server
									mainError = err
								}
								return
							}
						}
					}
				}
				mainError = fmt.Errorf("the timeout for node %s expired", host)
			} else {
				mainError = err
			}
		} else {
			mainError = err
		}
	}

	for _, server := range Env.Servers { // Start the pxeboot
		go pxeboot(server.node)
	}

	wg.Wait()

	if mainError != nil {
		msg := fmt.Sprintf("%v", mainError)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

}
