package main

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/pcc-blackbox/models"
	"github.com/platinasystems/test"
)

var PxeBootSelectedNodeId uint64
var hwInventory []pcc.HardwareInventory

func testHardwareInventory(t *testing.T) {
	t.Run("pxebootNode", pxebootNode)
	t.Run("checkNodeAdd", checkNodeAdd)
	t.Run("checkHardwareInventory", checkHardWareInventory)
	t.Run("checkStorage", checkStorage)
	t.Run("powerCycleNode", powerCycleNode)
}

func pxebootNode(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res, Env.CheckServers, CheckNodes)

	assert := test.Assert{t}
	var (
		err     error
		pxeboot []byte
	)
	if len(Env.Servers) != 0 {
		pxeboot, err = exec.Command("/bin/bash", "-c", "for cmd in 'chassis bootdev pxe' 'chassis power cycle'; do ipmitool -I lanplus -H "+Env.Servers[0].BMCIp+" -U ADMIN -P ADMIN $cmd; done").Output()
		if err != nil {
			msg := fmt.Sprintf("%v%v", string(pxeboot), err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
		}
	}
}

func checkNodeAdd(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	from := time.Now()
	err := verifyAddNode(from, "nodeAdd")
	if err != nil {
		msg := fmt.Sprintf("Node additon failed..ERROR:%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
}

func checkHardWareInventory(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res, CheckNodes)

	assert := test.Assert{t}
	var (
		flag bool
		err  error
	)

	hwInventory, err = Pcc.GetHardwareInventory()
	if err != nil {
		msg := fmt.Sprintf("GetHardwareInventory failed: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	for _, server := range Env.Servers {
		for _, hw := range hwInventory {
			if server.BMCIp == hw.Bus.Bmc.Ipcfg.Ipaddress {
				PxeBootSelectedNodeId = hw.NodeID
				flag = true
				log.AuctaLogger.Infof("Hardware inventory with node id %v persisted succesfully", PxeBootSelectedNodeId)
				break
			}
		}
	}
	if !flag {
		msg := fmt.Sprintf("inventory for node with id %v not persisted in db", PxeBootSelectedNodeId)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
}

func checkStorage(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	assert := test.Assert{t}

	var (
		storage pcc.StorageChildrenTO
		err     error
	)

	storage, err = Pcc.GetStorageNode(PxeBootSelectedNodeId)
	if err != nil {
		msg := fmt.Sprintf("GetStorageNode failed: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	if len(storage.Children) != 0 {
		log.AuctaLogger.Infof("inventory for node with id %v persisted in "+
			"storage inventory", PxeBootSelectedNodeId)
	} else {
		msg := fmt.Sprintf("inventory for node with id %v not persisted "+
			"in storage inventory", PxeBootSelectedNodeId)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
}

func powerCycleNode(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	assert := test.Assert{t}
	var (
		err        error
		powerCycle []byte
	)
	if len(Env.Servers) != 0 {
		powerCycle, err = exec.Command("/bin/bash", "-c", "ipmitool -I lanplus -H "+Env.Servers[0].BMCIp+" -U ADMIN -P ADMIN chassis power cycle").Output()
		if err != nil {
			msg := fmt.Sprintf("power cycle failed %v%v", string(powerCycle), err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
		}
	}

}

func verifyAddNode(from time.Time, action string) (err error) {
	done := make(chan status)
	breakLoop := make([]chan bool, 2)

	if action == "nodeAdd" {
		go func() {
			breakLoop[0] = make(chan bool)
			syncCheckGenericInstallation(0, PXEBOOT_TIMEOUT, PXEBOOT_NODE_ADD_NOTIFICATION, from, done, breakLoop[0])
		}()
		go func() {
			breakLoop[1] = make(chan bool)
			syncCheckGenericInstallation(0, PXEBOOT_TIMEOUT, PXEBOOT_NODE_ADD_FAILED_NOTIFICATION, from, done, breakLoop[1])
		}()
	}
	s := <-done
	go func() {
		for i := 0; i < 2; i++ {
			breakLoop[i] <- true
		}
	}()
	if !s.isError {
		if strings.Contains(s.msg, PXEBOOT_NODE_ADD_NOTIFICATION) {
			fmt.Println("Node is added succesfully..", s.msg)
		} else if strings.Contains(s.msg, PXEBOOT_NODE_ADD_FAILED_NOTIFICATION) {
			err = fmt.Errorf("%v", s.msg)
		}
	} else {
		err = fmt.Errorf("%v", s.msg)
	}
	return
}
