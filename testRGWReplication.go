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
	t.Run("downloadTrustCertificateSecondary", downloadTrustCertificateSecondary)
	t.Run("primarySideEndedTrustCreation", primarySideEndedTrustCreation)
	t.Run("primarySelectTargetNodes", primarySelectTargetNodes)
	t.Run("checkTrustResult", checkTrustResult)
	t.Run("deleteTrustSecondary", deleteTrustSecondary)
}

func testRGWReplicationPrimaryStarted(t *testing.T) {
	t.Run("initPccs", initPccs)
	t.Run("primarySideStartedTrustCreation", primarySideStartedTrustCreation)
	t.Run("downloadTrustCertificatePrimary", downloadTrustCertificatePrimary)
	t.Run("secondarySideEndedTrustCreation", secondarySideEndedTrustCreation)
	t.Run("primarySelectTargetNodes", primarySelectTargetNodes)
	t.Run("checkTrustResult", checkTrustResult)
	t.Run("deleteTrustPrimary", deleteTrustPrimary)
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

func primarySideStartedTrustCreation(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res)

	primaryRGW, err := PccPrimary.GetRadosGWByName(Env.RGWReplicationConfiguration.PrimaryRGWName)
	checkError(t, res, err)

	side := security.Master
	primaryTrust = security.Trust{
		Side:        &side,
		AppType:     "rgw",
		MasterAppID: primaryRGW.ID,
	}

	primaryTrust, err = PccPrimary.StartRemoteTrustCreation(&primaryTrust)
	checkError(t, res, err)

	log.AuctaLogger.Infof("Started trust creation: %v", primaryTrust)
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

func downloadTrustCertificatePrimary(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res)

	trustFile, err := PccPrimary.GetTrustFile(primaryTrust.ID)
	checkError(t, res, err)

	log.AuctaLogger.Infof("TrustFile response: %v", trustFile)

	token, err := json.MarshalIndent(trustFile, "", " ")
	err = ioutil.WriteFile("token.txt", token, 0644)
	checkError(t, res, err)
}

func downloadTrustCertificateSecondary(t *testing.T) {
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

func secondarySideEndedTrustCreation(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res)

	secondaryCluster, err := PccSecondary.GetCephCluster(Env.RGWReplicationConfiguration.SecondaryClusterName)
	checkError(t, res, err)

	var targetID uint64
	if Env.RGWReplicationConfiguration.TargetNodeName == "" {
		targetID = secondaryCluster.Nodes[0].NodeId
	} else {
		targetID, err = PccSecondary.FindNodeId(Env.RGWReplicationConfiguration.TargetNodeName)
		checkError(t, res, err)
	}

	params := pcc.SlaveParametersRGW{ClusterID: secondaryCluster.Id, TargetNodes: []uint64{targetID}}
	primaryTrust, err = PccPrimary.SecondaryEndedRemoteTrustCreation("rgw", params, "token.txt")
	checkError(t, res, err)
}

func primarySelectTargetNodes(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res)

	var err error
	var targetID uint64

	if Env.RGWReplicationConfiguration.TargetNodeName == "" {
		var params pcc.SlaveParametersRGW
		strParams, _ := json.Marshal(secondaryTrust.SlaveParams)
		json.Unmarshal(strParams, &params)

		targetID = params.AvailableNodes[0].ID
	} else {
		targetID, err = PccSecondary.FindNodeId(Env.RGWReplicationConfiguration.TargetNodeName)
		checkError(t, res, err)
	}

	trust := security.Trust{
		ID:          primaryTrust.ID,
		SlaveParams: map[string]interface{}{"targetNodes": []uint64{targetID}},
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
	secondaryTrust, err = PccSecondary.GetTrust(secondaryTrust.ID)
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
			secondaryTrust, err = PccSecondary.GetTrust(secondaryTrust.ID)
			checkError(t, res, err)

			if primaryTrust.Status == security.SetUp && secondaryTrust.Status == security.SetUp {
				log.AuctaLogger.Info("Trust has been established")
				return
			} else if primaryTrust.Status == security.Ready && secondaryTrust.Status == security.Ready {
				err = errors.New("Trust setup failed")
				checkError(t, res, err)
			} else {
				log.AuctaLogger.Infof("Primary trust status: %s", primaryTrust.Status)
				log.AuctaLogger.Infof("Secondary trust status: %s", secondaryTrust.Status)
			}
		}
	}
}

func deleteTrustPrimary(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res)

	_, err := PccPrimary.DeleteTrust(primaryTrust.ID)
	checkError(t, res, err)

	var primaryTrustExists, secondaryTrustExists bool

	timeout := time.After(15 * time.Minute)
	tick := time.Tick(30 * time.Second)
	for true {
		select {
		case <-timeout:
			msg := "Timed out waiting for trust to be established"
			checkError(t, res, errors.New(msg))
		case <-tick:
			primaryTrustExists, err = PccPrimary.TrustExists(primaryTrust.ID)
			checkError(t, res, err)
			secondaryTrustExists, err = PccSecondary.TrustExists(secondaryTrust.ID)
			checkError(t, res, err)

			if !primaryTrustExists && !secondaryTrustExists {
				log.AuctaLogger.Info("Trust has been revoked")
				return
			} else {
				log.AuctaLogger.Infof("Primary trust exists: %s", primaryTrustExists)
				log.AuctaLogger.Infof("Secondary trust exists: %s", secondaryTrustExists)
			}
		}
	}
}

func deleteTrustSecondary(t *testing.T) {
	test.SkipIfDryRun(t)

	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res)

	_, err := PccSecondary.DeleteTrust(secondaryTrust.ID)
	checkError(t, res, err)

	var primaryTrustExists, secondaryTrustExists bool

	timeout := time.After(15 * time.Minute)
	tick := time.Tick(30 * time.Second)
	for true {
		select {
		case <-timeout:
			msg := "Timed out waiting for trust to be established"
			checkError(t, res, errors.New(msg))
		case <-tick:
			primaryTrustExists, err = PccPrimary.TrustExists(primaryTrust.ID)
			checkError(t, res, err)
			secondaryTrustExists, err = PccSecondary.TrustExists(secondaryTrust.ID)
			checkError(t, res, err)

			if !primaryTrustExists && !secondaryTrustExists {
				log.AuctaLogger.Info("Trust has been revoked")
				return
			} else {
				log.AuctaLogger.Infof("Primary trust exists: %s", primaryTrustExists)
				log.AuctaLogger.Infof("Secondary trust exists: %s", secondaryTrustExists)
			}
		}
	}
}
