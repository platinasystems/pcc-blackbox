package models

import (
	"testing"
	"time"

	db "github.com/platinasystems/go-common/database"
	log "github.com/platinasystems/go-common/logs"
	"github.com/platinasystems/pcc-blackbox/utility"
)

const (
	TestPass      string = "pass"
	TestFail      string = "fail"
	TestUndefined string = "undefined"
)

type TestResult struct {
	ID          uint
	RunID       string
	TestID      string
	Result      string
	FailureMsg  string
	ElapsedTime time.Duration
}

func InitTestResult(runID string) (res *TestResult) {
	res = &TestResult{RunID: runID, Result: TestUndefined}
	res.TestID = utility.FuncName()
	return
}

func (testResult *TestResult) SetTestFailure(failureMsg string) {
	testResult.Result = TestFail
	testResult.FailureMsg = failureMsg
}

func (testResult *TestResult) SetTestPass() {
	testResult.Result = TestPass
}

func (testResult *TestResult) SetElapsedTime(start time.Time, name string) {
	testResult.ElapsedTime = time.Since(start)
	log.AuctaLogger.Infof("%s took %s", name, testResult.ElapsedTime)
}

func (testResult *TestResult) SaveTestResult() {
	db.NewDBHandler().GetDM().Create(testResult)
}

func (testResult *TestResult) CheckTestAndSave(t *testing.T, start time.Time, funcName string) {
	if !t.Failed() {
		testResult.SetTestPass()
	}
	testResult.SetElapsedTime(start, funcName)
	testResult.SaveTestResult()
}
