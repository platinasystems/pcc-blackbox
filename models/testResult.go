package models

import (
	db "github.com/platinasystems/go-common/database"
	log "github.com/platinasystems/go-common/logs"
	"github.com/platinasystems/pcc-blackbox/utility"
	"time"
)

const (
	TestPass string = "pass"
	TestFail string = "fail"
)

type TestResult struct {
	ID          uint
	RunID       string
	TestID      uint32
	Result      string
	FailureMsg  string
	ElapsedTime time.Duration
}

func InitTestResult(runID string) (res *TestResult) {
	res = &TestResult{RunID: runID}
	res.TestID = utility.GetTestID()
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
