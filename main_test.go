// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime/debug"
	"testing"
	"time"

	"github.com/google/uuid"
	db "github.com/platinasystems/go-common/database"
	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/pcc-blackbox/models"
	"github.com/platinasystems/pcc-blackbox/utility"
	"github.com/platinasystems/test"
)

var envFile string = "testEnv.json"
var customTestsFileName string
var seed int64

func init() {
	flag.Int64Var(&seed, "seed", -1, "seed to initialize random generator")
	flag.StringVar(&customTestsFileName, "testfile", "testList.yml.example", "name of the file with a custom list of test")

	nameToTestFunc = GetNameToTestFunc()
	defaultTests = GetDefaultTestMap()

}

func TestMain(m *testing.M) {
	var (
		ecode int
		err   error
	)
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, r)
			ecode = 1
			log.AuctaLogger.Info(string(debug.Stack()))
		}
		if ecode != 0 {
			os.Exit(ecode)
		}
	}()

	if _, err := os.Stat("logConfig.yml"); err == nil {
		pcc.LoadLogConfig("logConfig.yml", "yml")
		log.Init()
	} else if os.IsNotExist(err) {
		log.InitWithDefault(nil)
	}

	data, err := ioutil.ReadFile(envFile)
	if err != nil {
		panic(fmt.Errorf("Error opening %v: %v", envFile, err))
	}

	if err = json.Unmarshal(data, &Env); err != nil {
		if jsonErr, ok := err.(*json.SyntaxError); ok {
			// emacs users can use M-x goto-char <offset>
			part := data[jsonErr.Offset-10 : jsonErr.Offset+10]
			err = fmt.Errorf("%w ~ error near '%s' (offset %d)",
				err, part, jsonErr.Offset)
		}
		panic(fmt.Errorf("error unmarshalling %v: %v",
			envFile, err.Error()))
	}

	params := &db.Params{DBtype: "sqlite3", DBname: "blackbox.db"}
	dbh := db.InitWithParams(params, false)
	if dbh == nil {
		log.AuctaLogger.Errorf("No database handler initialized")
	} else {
		models.DBh = dbh
		dbh.GetDM().AutoMigrate(&models.TestResult{})
		dbh.GetDM().AutoMigrate(&models.RandomSeed{})
	}

	credential := pcc.Credential{ // FIXME move to json
		UserName: "admin",
		Password: "admin",
	}
	if Pcc, err = pcc.Init(Env.PccIp, credential, Env.DBConfiguration, Env.SshConfiguration); err != nil {
		panic(fmt.Errorf("Authentication error: %v", err))
	}

	dockerStats = pcc.InitDockerStats(Env.DockerStats)
	err = pcc.StoreContainerNames()
	if err != nil {
		fmt.Printf("Error storing containers: %v", err)
	}
	flag.Parse()
	if *test.DryRun {
		m.Run()
		return
	}

	runID = uuid.New().String()
	log.AuctaLogger.Infof("Generated runID: %s", runID)

	if seed == -1 {
		seed = utility.CreateSeed()
	}
	randomGenerator = utility.RandomGenerator(seed)
	randomSeed := models.RandomSeed{RunID: runID, Seed: seed}
	if dbh != nil {
		dbh.Insert(&randomSeed)
	}

	startTime := ConvertToMillis(time.Now())
	ecode = m.Run()
	stopTime := ConvertToMillis(time.Now())

	dockerStats.Stop()
	utility.SaveNodesHistoricalSummaries(Pcc, runID, startTime, stopTime)
	log.AuctaLogger.Info("TEST COMPLETED")
	log.AuctaLogger.Flush()
}

var count uint
var timeFormat = "Mon Jan 2 15:04:05 2006"

// TestNodes can be used to
// automatically config a cluser
func TestNodes(t *testing.T) {
	count++
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "nodes", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "getSecKeys", getSecKeys)
		mayRun(t, "updateSecurityKey", updateSecurityKey_MaaS)
		mayRun(t, "addInvaders", addClusterHeads)
		mayRun(t, "addBrownfieldNodes", addBrownfieldServers)
		mayRun(t, "configServerInterfaces", configServerInterfaces)
		mayRun(t, "updateBmcInfo", updateBmcInfo)
	})
}

func TestUsers(t *testing.T) {
	count++
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "users", func(t *testing.T) {
		mayRun(t, "addTenant", addTenant)
	})
}

// assumes TestNodes has been run
func TestMaaS(t *testing.T) {
	count++
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "nodes", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "getSecKeys", getSecKeys)
		mayRun(t, "updateSecurityKey", updateSecurityKey_MaaS)
		mayRun(t, "installMAAS", updateNodes_installMAAS)
		mayRun(t, "reimageAllBrownNodes", reimageAllBrownNodes)
	})
}

// assumes TestNodes has been run
func TestTenantMaaS(t *testing.T) {
	count++
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "nodes", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "getSecKeys", getSecKeys)
		mayRun(t, "updateSecurityKey", updateSecurityKey_MaaS)
		mayRun(t, "installMAAS", updateNodes_installMAAS)
		mayRun(t, "addTenant", addTenant)
		mayRun(t, "reimageTenantAllBrownNodes", reimageAllBrownNodes)
	})
}

// assumes TestNode & TestNetCluster has been run before
func TestAddK8s(t *testing.T) {
	count++
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "nodes", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "addIpam", updateIpam)
		mayRun(t, "addNetCluster", addNetCluster)
		mayRun(t, "CreateK8sCluster", createK8sCluster)
	})
}

func TestDeleteK8s(t *testing.T) {
	count++
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "nodes", func(t *testing.T) {
		mayRun(t, "deleteK8sCluster", deleteK8sCluster)
	})
}

// assumes TestNodes has been run
func TestNetCluster(t *testing.T) {
	count++
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "netCluster", func(t *testing.T) {
		mayRun(t, "getNodesList", getNodes)
		mayRun(t, "addIpam", updateIpam)
		mayRun(t, "addNetCluster", addNetCluster)
	})
}

// assumes TestNode & TestNetCluster has been run before
func TestAddCeph(t *testing.T) {
	count++
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "ceph", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "addIpam", updateIpam)
		mayRun(t, "addNetCluster", addNetCluster)
		mayRun(t, "testCeph", testCeph)
	})
}

// assumes TestAddCeph has been run
func TestDeleteCeph(t *testing.T) {
	count++
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "cephDelete", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "testDeleteCeph", testDeleteCeph)
	})
}

// assumes TestAddCeph has been run
func TestCephCache(t *testing.T) {
	mayRun(t, "ceph-cache", func(t *testing.T) {
		mayRun(t, "testCephCacheSetup", testCephCacheSetup)
		mayRun(t, "testCephCacheAdd", testCephCacheAdd)
		mayRun(t, "testCephCacheDelete", testCephCacheDelete)
	})
}

func TestK8sApp(t *testing.T) {
	count++
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "K8sApp", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "addIpam", updateIpam)
		mayRun(t, "addNetCluster", addNetCluster)
		mayRun(t, "testK8sApp", testK8sApp)
	})
}

func TestPortus(t *testing.T) {
	count++
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "portus", func(t *testing.T) {
		mayRun(t, "getNodesList", getNodes)
		mayRun(t, "addBrownfieldNodes", addBrownfieldServers)
		mayRun(t, "uploadSecurityAuthProfileCertificate", UploadSecurityAuthProfileCert)
		mayRun(t, "addProfile", AddAuthenticationProfile)
		mayRun(t, "uploadSecurityPortusKey", UploadSecurityPortusKey)
		mayRun(t, "uploadSecurityPortusCertificate", UploadSecurityPortusCert)
		mayRun(t, "installPortus", AddPortus)
		mayRun(t, "checkPortusInstallation", CheckPortusInstallation)
	})
}

// assumes TestPortus has been run
func TestDelPortus(t *testing.T) {
	count++
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "portus", func(t *testing.T) {
		mayRun(t, "getAvailableNodes", getAvailableNodes)
		mayRun(t, "delAllPortus", delAllPortus)
	})
}

// assumes TestNodes has been run
// requires ipmitool installed on unit running test
func TestHardwareInventory(t *testing.T) {
	count++
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "hardwareinventory", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "testHardwareInventory", testHardwareInventory)
	})
}

func TestFull(t *testing.T) {
	count++
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "nodes", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "getSecKeys", getSecKeys)
		mayRun(t, "updateSecurityKey", updateSecurityKey_MaaS)
		mayRun(t, "addInvaders", addClusterHeads)
		mayRun(t, "addBrownfieldNodes", addBrownfieldServers)
		mayRun(t, "installLLDP", updateNodes_installLLDP)
		mayRun(t, "installMAAS", updateNodes_installMAAS)
		mayRun(t, "configServerInterfaces", configServerInterfaces)
		mayRun(t, "reimageAllBrownNodes", reimageAllBrownNodes)
		mayRun(t, "addTenant", addTenant)
		mayRun(t, "reimageTenantAllBrownNodes", reimageAllBrownNodes)
		mayRun(t, "CreateK8sCluster", createK8sCluster)
	})
}

func TestClean(t *testing.T) {
	count++
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "nodes", func(t *testing.T) {
		mayRun(t, "getAvailableNodes", getAvailableNodes)
		mayRun(t, "deleteK8sCluster", deleteK8sCluster)
		mayRun(t, "delAllNetsCluster", delAllNetsCluster)
		mayRun(t, "delAllPortus", delAllPortus)
		mayRun(t, "delAllNodes", delAllNodes)
		mayRun(t, "delAllUsers", delAllUsers)
		mayRun(t, "delAllTenants", delAllTenants)
		mayRun(t, "delAllProfiles", delAllProfiles)
		mayRun(t, "delAllCerts", delAllCerts)
		mayRun(t, "delAllKeys", delAllKeys)
		mayRun(t, "delAllIpams", delAllIpams)
	})
}

func TestTunnel(t *testing.T) {
	count++
	log.AuctaLogger.Info("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "TUNNEL", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "addInvaders", addClusterHeads)
		mayRun(t, "addBrownfieldNodes", addBrownfieldServers)
		mayRun(t, "checkInvaderTunnels", checkInvaderTunnels)
		mayRun(t, "checkServerTunnels", checkServerTunnels)
		mayRun(t, "checkTunnelConnection", checkTunnelConnection)
		mayRun(t, "checkTunnelForwardConnection", checkTunnelForwardingRules)
		mayRun(t, "checkTunnelRestoreConnection", checkTunnelRestore)
	})
}

func TestPolicy(t *testing.T) {
	count++
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "POLICY", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "addInvaders", addClusterHeads)
		mayRun(t, "testPreparePolicies", testPreparePolicies)
		mayRun(t, "testPolicies", testPolicies)
		mayRun(t, "testPolicyScope", testPolicyScope)
	})
}

func TestAvailability(t *testing.T) {
	count++
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "AVAILABILITY", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "addInvaders", addClusterHeads)
		mayRun(t, "checkAddUnreachableNode", addUnreachableNode)
		mayRun(t, "checkAddInaccessibleNode", addInaccessibleNode)
		mayRun(t, "checkAgentAndCollectorRestore", checkAgentAndCollectorRestore)
	})
}

// assumes TestNodes has been run
func TestGreenfield(t *testing.T) {
	count++
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "GREENFIELD", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "updateSecurityKey", updateSecurityKey_MaaS)
		mayRun(t, "installMAAS", installMAAS)
		mayRun(t, "addGreenfieldNodes", addGreenfieldServers)
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "configNetworkInterfaces", configNetworkInterfaces)
		mayRun(t, "reimageAllBrownNodes", reimageAllBrownNodes)
	})
}

func TestConfigNetworkInterfaces(t *testing.T) {
	count++
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "configureNetwork", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "configNetworkInterfaces", configNetworkInterfaces)
	})
}

func TestMonitor(t *testing.T) {
	count++
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "MONITOR", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "addInvaders", addClusterHeads)
		mayRun(t, "testGetTopic", testGetTopic)
		mayRun(t, "testGetTopicSchema", testGetTopicSchema)
		mayRun(t, "testMonitorSample", testMonitorSample)
		mayRun(t, "testMonitorHistory", testMonitorHistory)
	})
}

func TestUserManagement(t *testing.T) {
	count++
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "USER-MANAGEMENT", func(t *testing.T) {
		mayRun(t, "testUMRole", testUMRole)
		mayRun(t, "testUMTenant", testUMTenant)
		mayRun(t, "testUMUser", testUMUser)
		mayRun(t, "testUMOperation", testUMOperation)
		mayRun(t, "testUMEntity", testUMEntity)
		mayRun(t, "testUMUserSpace", testUMUserSpace)
	})
}

func TestKeyManager(t *testing.T) {
	count++
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "KEY-MANAGER", func(t *testing.T) {
		mayRun(t, "testKMKeys", testKMKeys)
		mayRun(t, "testKMCertificates", testKMCertificates)
	})
}

func TestAppCredentials(t *testing.T) {
	count++
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "APP CREDENTIALS", func(t *testing.T) {
		mayRun(t, "testCreateCredendialMetadataProfile", testCreateCredendialMetadataProfile)
		mayRun(t, "testUpdateCredendialMetadataProfile", testUpdateCredendialMetadataProfile)
		mayRun(t, "testDeleteCredendialMetadataProfile", testDeleteCredendialMetadataProfile)
	})
}

// Test functions for New Dashboard
func TestDashboard(t *testing.T) {
	count++
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))
	mayRun(t, "DASHBOARD REST API", func(t *testing.T) {
		mayRun(t, "testDashboardGetAllPCCObjects", testDashboardGetAllPCCObjects)
		mayRun(t, "testDashboardGetPCCObjectByRandomId", testDashboardGetPCCObjectByRandomId)
		// ** calling with a fixed id (hardcoded to 4 in the code, may return a Not Found
		// ** Hence comment this test, as it is semantically equivalent to the objectByRandomId call above
		// ** mayRun(t, "testDashboardGetPCCObjectById", testDashboardGetPCCObjectById)
		mayRun(t, "testDashboardGetChildrenObjectsByRandomId", testDashboardGetChildrenObjectsByRandomId)
		mayRun(t, "testDashboardGetParentObjectsByRandomId", testDashboardGetParentObjectsByRandomId)
		mayRun(t, "testDashboardGetFilteredObjects", testDashboardGetFilteredObjects)
		mayRun(t, "testDashboardGetAdvSearchedObjects", testDashboardGetAdvSearchedObjects)
		mayRun(t, "testDashboardGetAggrHealthCountByType", testDashboardGetAggrHealthCountByType)
		mayRun(t, "testDashboardGetMetadataEnumStrings", testDashboardGetMetadataEnumStrings)
	})
}

func TestCustom(t *testing.T) {
	log.AuctaLogger.Infof("Iteration %v, %v", count, time.Now().Format(timeFormat))

	customTests, err := utility.GetCustomTests(customTestsFileName)
	if err != nil {
		log.AuctaLogger.Errorf("%v", err)
		t.SkipNow()
	}

	for testName, subtests := range customTests.TestList {
		count++

		if defaultSubtests, ok := defaultTests[testName]; ok {
			subtests = defaultSubtests
		}

		for _, subtest := range subtests {
			if _, ok := nameToTestFunc[subtest]; !ok {
				log.AuctaLogger.Errorf("There is no function named %s", subtest)
				t.SkipNow()
			}
		}

		mayRun(t, testName, func(t *testing.T) {
			for _, subtest := range subtests {
				mayRun(t, subtest, nameToTestFunc[subtest])
			}
		})
	}
}

func TestGen(t *testing.T) {
	// Not a real testcase, but can be used to generate a
	// testEnv.json file from existing PCC setup.
	test.SkipIfDryRun(t)
	genEnv()
	os.Exit(0)
}

func mayRun(t *testing.T, name string, f func(*testing.T)) bool {
	dockerStats.ChangePhase(name)
	var ret bool
	t.Helper()
	if !t.Failed() {
		ret = t.Run(name, f)
	}
	return ret
}

func uutInfo() {
	log.AuctaLogger.Info("---")
	defer log.AuctaLogger.Info("...")
	log.AuctaLogger.Info("pcc instance unknown")
}
