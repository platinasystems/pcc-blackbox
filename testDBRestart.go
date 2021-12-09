package main

import (
	"errors"
	"fmt"
	"github.com/platinasystems/pcc-blackbox/models"
	"github.com/platinasystems/test"
	"os/exec"
	"reflect"
	"testing"
	"time"
)

const (
	StartCommand            = "start"
	StopCommand             = "stop"
	RestartCommand          = "restart"
	DBContainer             = "postgres-db"
	PCCServerContainer      = "pccserver"
	KeyManagerContainer     = "key-manager"
	UserManagementContainer = "user-management"
	PlatinaMonitorContainer = "platina-monitor"
	SecurityContainer       = "security"
)

func testDBRestartPCCServer(t *testing.T) {
	t.Run("stopDB", stopDB)
	t.Run("restartPCCServer", restartPCCServer)

	time.Sleep(3 * time.Minute)

	t.Run("startDB", startDB)
	t.Run("checkPCCServerStatus", checkPCCServerStatus)
}

func testDBRestartUserManagement(t *testing.T) {
	t.Run("stopDB", stopDB)
	t.Run("restartUserManagement", restartUserManagement)

	time.Sleep(3 * time.Minute)

	t.Run("startDB", startDB)
	t.Run("checkUserManagementStatus", checkUserManagementStatus)
}

func testDBRestartPlatinaMonitor(t *testing.T) {
	t.Run("stopDB", stopDB)
	t.Run("restartPlatinaMonitor", restartPlatinaMonitor)

	time.Sleep(3 * time.Minute)

	t.Run("startDB", startDB)
	t.Run("checkPlatinaMonitorStatus", checkPlatinaMonitorStatus)
}

func testDBRestartSecurity(t *testing.T) {
	t.Run("stopDB", stopDB)
	t.Run("restartSecurity", restartSecurity)

	time.Sleep(3 * time.Minute)

	t.Run("startDB", startDB)
	t.Run("checkSecurityStatus", checkSecurityStatus)
}

func testDBRestartKeyManager(t *testing.T) {
	t.Run("stopDB", stopDB)
	t.Run("restartKeyManager", restartKeyManager)

	time.Sleep(3 * time.Minute)

	t.Run("startDB", startDB)
	t.Run("checkKeyManagerStatus", checkKeyManagerStatus)
}

func opContainer(t *testing.T, res *models.TestResult, containerName string, command string) {
	cmd := exec.Command("docker", command, containerName)
	err := cmd.Run()
	checkError(t, res, err)
}

func stopDB(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	opContainer(t, res, DBContainer, StopCommand)
}

func startDB(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	opContainer(t, res, DBContainer, StartCommand)
}

func restartPCCServer(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	opContainer(t, res, PCCServerContainer, RestartCommand)
}

func restartKeyManager(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	opContainer(t, res, KeyManagerContainer, RestartCommand)
}

func restartUserManagement(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	opContainer(t, res, UserManagementContainer, RestartCommand)
}

func restartSecurity(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	opContainer(t, res, SecurityContainer, RestartCommand)
}

func restartPlatinaMonitor(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	opContainer(t, res, PlatinaMonitorContainer, RestartCommand)
}

func checkPCCServerStatus(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	checkContainerStatus(t, res, PCCServerContainer, "GetNodes")
}

func checkKeyManagerStatus(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	checkContainerStatus(t, res, KeyManagerContainer, "GetSecurityKeys")
}

func checkUserManagementStatus(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	checkContainerStatus(t, res, UserManagementContainer, "GetUsers")
}

func checkPlatinaMonitorStatus(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	checkContainerStatus(t, res, PlatinaMonitorContainer, "GetTopics")
}

func checkSecurityStatus(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	timeout := time.After(3 * time.Minute)
	tick := time.Tick(30 * time.Second)
	for true {
		select {
		case <-timeout:
			err := errors.New(fmt.Sprintf("Timeout waiting for %s", SecurityContainer))
			checkError(t, res, err)
		case <-tick:
			err := Pcc.ChangeUser(adminCredential)
			if err == nil {
				return
			}
		}
	}
}

func checkContainerStatus(t *testing.T, res *models.TestResult, containerName string, testFunction string) {
	timeout := time.After(3 * time.Minute)
	tick := time.Tick(30 * time.Second)
	for true {
		select {
		case <-timeout:
			err := errors.New(fmt.Sprintf("Timeout waiting for %s", containerName))
			checkError(t, res, err)
		case <-tick:
			result := reflect.ValueOf(Pcc).MethodByName(testFunction).Call(nil)
			err := result[1].Interface()
			if err == nil {
				return
			}
		}
	}
}
