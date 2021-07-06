package main

import (
	"encoding/json"
	"errors"
	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	m "github.com/platinasystems/pcc-blackbox/models"
	"github.com/platinasystems/pcc-models/security"
	"github.com/platinasystems/test"
	"io/ioutil"
	"testing"
	"time"
)

var (
	PccPrimary     *pcc.PccClient
	PccSecondary   *pcc.PccClient
	primaryTrust   security.Trust
	secondaryTrust security.Trust
)

func testRGWReplicationSecondaryStarted(t *testing.T) {
	t.Run("initPccs", initPccs)
	t.Run("secondarySideStartedTrustCreation", secondarySideStartedTrustCreation)
	t.Run("downloadTrustCertificate", downloadTrustCertificate)
	t.Run("primarySideEndedTrustCreation", primarySideEndedTrustCreation)
	t.Run("primarySelectTargetNodes", primarySelectTargetNodes)
	t.Run("checkTrustResult", checkTrustResult)
}

func initPccs(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res)

	var err error
	PccPrimary, err = pcc.Init(Env.RGWReplicationConfiguration.PccPrimaryIP, adminCredential, nil, nil)
	checkError(t, res, err)
	PccSecondary, err = pcc.Init(Env.RGWReplicationConfiguration.PccSecondaryIP, adminCredential, nil, nil)
	checkError(t, res, err)
}

func secondarySideStartedTrustCreation(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res)

	secondaryCluster, err := PccSecondary.GetCephCluster(Env.RGWReplicationConfiguration.SecondaryClusterName)
	checkError(t, res, err)

	side := security.Slave
	secondaryTrust = security.Trust{
		Side:        &side,
		AppType:     "rgw",
		SlaveParams: map[string]interface{}{"clusterID": secondaryCluster.Id},
	}

	secondaryTrust, err = PccSecondary.StartRemoteTrustCreation(&secondaryTrust)
	checkError(t, res, err)

	log.AuctaLogger.Infof("Started trust creation: %v", secondaryTrust)
}
func downloadTrustCertificate(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res)

	trustFile, err := PccSecondary.GetTrustFile(secondaryTrust.ID)
	checkError(t, res, err)

	log.AuctaLogger.Infof("TrustFile response: %v", trustFile)

	token, err := json.MarshalIndent(trustFile, "", " ")
	err = ioutil.WriteFile("token.txt", token, 0644)
	checkError(t, res, err)
}

func primarySideEndedTrustCreation(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res)

	primaryRGW, err := PccPrimary.GetRadosGWByName(Env.RGWReplicationConfiguration.PrimaryRGWName)
	checkError(t, res, err)

	primaryTrust, err = PccPrimary.PrimaryEndedRemoteTrustCreation("rgw", primaryRGW.ID, "token.txt")
	checkError(t, res, err)
}

func primarySelectTargetNodes(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res)

	var err error
	var params pcc.SlaveParametersRGW
	strParams, _ := json.Marshal(secondaryTrust.SlaveParams)
	json.Unmarshal(strParams, &params)

	log.AuctaLogger.Info(strParams)
	log.AuctaLogger.Info(params)

	trust := security.Trust{
		ID:          primaryTrust.ID,
		SlaveParams: map[string]interface{}{"targetNodes": []uint64{params.AvailableNodes[0].ID}},
	}

	primaryTrust, err = PccPrimary.SelectTargetNodes(&trust, primaryTrust.ID)
	checkError(t, res, err)

}

func checkTrustResult(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res)

	var err error
	primaryTrust, err = PccPrimary.GetTrust(primaryTrust.ID)
	checkError(t, res, err)
	secondaryTrust, err = PccPrimary.GetTrust(secondaryTrust.ID)
	checkError(t, res, err)

	timeout := time.After(15 * time.Minute)
	tick := time.Tick(30 * time.Second)
	for true {
		select {
		case <-timeout:
			msg := "Timed out waiting for trust to be established"
			checkError(t, res, errors.New(msg))
		case <-tick:
			primaryTrust, err = PccPrimary.GetTrust(primaryTrust.ID)
			checkError(t, res, err)
			secondaryTrust, err = PccPrimary.GetTrust(secondaryTrust.ID)
			checkError(t, res, err)

			if primaryTrust.Status == security.SetUp && secondaryTrust.Status == security.SetUp {
				log.AuctaLogger.Info("Trust has been established")
				return
			} else {
				log.AuctaLogger.Infof("Primary trust status: %s", primaryTrust.Status)
				log.AuctaLogger.Infof("Secondary trust status: %s", secondaryTrust.Status)
			}
		}
	}
}
