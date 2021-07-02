package main

import (
	m "github.com/platinasystems/pcc-blackbox/models"
	"github.com/platinasystems/test"
	"testing"
	"time"
)

func testRGWReplicationSecondaryStarted(t *testing.T) {
	t.Run("secondarySideTrustCreation", secondarySideStartedTrustCreation)
	t.Run("primarySideTrustCreation", primarySideEndedTrustCreation)
	t.Run("primarySelectTargetNodes", primarySelectTargetNodes)
	t.Run("checkTrustResult", checkTrustResult)
}

func secondarySideStartedTrustCreation(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res)
}

func primarySideEndedTrustCreation(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res)
}

func primarySelectTargetNodes(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res)
}

func checkTrustResult(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res)
}
