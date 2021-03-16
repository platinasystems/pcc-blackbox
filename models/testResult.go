package models

import (
	"testing"
	"time"

	log "github.com/platinasystems/go-common/logs"
	"github.com/platinasystems/pcc-blackbox/utility"
)

const (
	TestPass      string = "pass"
	TestFail      string = "fail"
	TestSkipped   string = "skipped"
	TestUndefined string = "undefined"
)

type TestResult struct {
	ID          uint
	RunID       string
	TestID      string
	Result      string
	FailureMsg  string
	ElapsedTime float64
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

func (testResult *TestResult) SetTestSkipped(failureMsg string) {
	testResult.FailureMsg = failureMsg
	testResult.Result = TestSkipped
}

func (testResult *TestResult) SetElapsedTime(start time.Time, name string) {
	testResult.ElapsedTime = time.Since(start).Seconds()
	log.AuctaLogger.Infof("%s took %fs", name, testResult.ElapsedTime)
}

func (testResult *TestResult) SaveTestResult() {
	if DBh != nil {
		DBh.Insert(testResult)
	} else {
		log.AuctaLogger.Error("Cannot save test result: No db handler initialized")
	}
}

func (testResult *TestResult) CheckTestAndSave(t *testing.T, start time.Time) {
	if !t.Failed() && !t.Skipped() {
		testResult.SetTestPass()
	}
	testResult.SetElapsedTime(start, testResult.TestID)
	testResult.SaveTestResult()
}
