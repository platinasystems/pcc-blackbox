package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/pcc-blackbox/models"
	"github.com/platinasystems/test"
)

const (
	KEYMANAGER_ENDPOINT  = "key-manager"
	PORTUS_KEY_FILENAME  = "test_portus_key"
	PORTUS_CERT_FILENAME = "test_portus_crt"
)

var PortusSelectedNodeIds []uint64

func AddPortus(t *testing.T) {
	t.Run("addPortus", installPortus)
}

func CheckPortusInstallation(t *testing.T) {
	t.Run("checkPortus", checkPortus)
}

func UploadSecurityPortusKey(t *testing.T) {
	t.Run("uploadPortusKey", uploadSecurityKey_Portus)
}

func UploadSecurityPortusCert(t *testing.T) {
	t.Run("uploadSecurityPortusCert", uploadCertificate_Portus)
}

func uploadSecurityKey_Portus(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now(), "uploadSecurityKey_Portus")

	assert := test.Assert{t}
	err := CreateFileAndUpload(PORTUS_KEY_FILENAME, PORTUS_KEY,
		pcc.PRIVATE_KEY, 0)
	if err != nil {
		msg := err.Error()
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
}

func uploadCertificate_Portus(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now(), "uploadCertificate_Portus")

	assert := test.Assert{t}
	var keyId uint64
	exist, privateKey, err := Pcc.FindSecurityKey(PORTUS_KEY_FILENAME)
	if err != nil {
		log.AuctaLogger.Errorf("Get private key %s failed\n%v\n", PORTUS_KEY_FILENAME, err)
	} else if exist {
		keyId = privateKey.Id
	}
	err = CreateFileAndUpload(PORTUS_CERT_FILENAME, PORTUS_CERT, pcc.CERT, keyId)
	if err != nil {
		msg := err.Error()
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
}

func installPortus(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now(), "installPortus")

	assert := test.Assert{t}
	var (
		portusConfiguration pcc.PortusConfiguration
	)

	for id, node := range Nodes {
		if Pcc.IsNodeOnline(node.Id) {
			portusConfiguration = Env.PortusConfiguration
			portusConfiguration.NodeID = id
			portusConfiguration.Name = fmt.Sprintf("portus_%v", id)

			if Env.AuthenticationProfile.Name == "" {
				log.AuctaLogger.Warnf("Authenticatiom Profile is not defined in the configuration file, Portus will be installed without it")
			} else {
				authProfile, err := Pcc.GetAuthProfileByName(CurrentAuthProfileName)
				if err == nil {
					portusConfiguration.AuthenticationProfileId = &authProfile.ID
				} else {
					log.AuctaLogger.Warnf("Missing authentication profile %s\n, Portus will be installed without it", CurrentAuthProfileName)
				}
			}

			exist, certificate, err := Pcc.FindCertificate(PORTUS_CERT_FILENAME)
			if err != nil {
				log.AuctaLogger.Errorf("Get certificate %s failed\n%v\n", PORTUS_CERT_FILENAME, err)
			} else if exist {
				portusConfiguration.RegistryCertId = &certificate.Id
			}

			log.AuctaLogger.Infof("Installing Portus on Node with id %v\n",
				node.Id)

			err = Pcc.InstallPortusNode(portusConfiguration)
			if err != nil {
				log.AuctaLogger.Warnf("Portus installation in %v failed\n%v\n", node.Host, err)
				log.AuctaLogger.Warnf("Trying in another node\n")
			} else {
				PortusSelectedNodeIds = append(PortusSelectedNodeIds, node.Id)
				break
			}
		}
	}
	if len(PortusSelectedNodeIds) == 0 {
		msg := "Failed to install Portus: No available nodes\n"
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
}

func checkPortus(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now(), "checkPortus")

	assert := test.Assert{t}

	for id, node := range Nodes {
		if idInSlice(node.Id, PortusSelectedNodeIds) {
			check, err := Pcc.WaitForInstallation(id, PORTUS_TIMEOUT, PORTUS_NOTIFICATION, "", nil)
			if err != nil {
				msg := fmt.Sprintf("Portus installation has failed\n%v\n", err)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				assert.FailNow()
			}
			if check {
				log.AuctaLogger.Infof("Portus correctly installed on nodeId:%v\n", id)
			}
		}
	}
}

func delAllPortus(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now(), "delAllPortus")

	assert := test.Assert{t}

	var (
		portusConfigs []pcc.PortusConfiguration
		err           error
		id            uint64
	)

	portusConfigs, err = Pcc.GetPortusNodes()
	if err != nil {
		msg := fmt.Sprintf("Failed to get portus nodes: %v\n", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}

	for _, p := range portusConfigs {
		log.AuctaLogger.Infof("Deleting Portus %v\n", p.Name)
		id = p.ID
		err = Pcc.DelPortusNode(id, true)
		if err != nil {
			msg := fmt.Sprintf("Failed to delete Portus %v: %v\n",
				p.Name, err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
			return
		}
		// wait till deleted
		done := false
		timeout := time.After(10 * time.Minute)
		tick := time.Tick(30 * time.Second)
		for !done {
			select {
			case <-tick:
				_, err = Pcc.GetPortusNodeById(id)
				if err != nil {
					if strings.Contains(err.Error(), "record not found") {
						done = true
						continue
					}
					msg := fmt.Sprintf("Failed Get Portus: %v\n",
						err)
					res.SetTestFailure(msg)
					log.AuctaLogger.Error(msg)
					assert.FailNow()
					return
				}
			case <-timeout:
				msg := fmt.Sprintf("Timeout deleting Portus\n")
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				assert.FailNow()
			}
		}
	}
}
