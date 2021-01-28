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

	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

var envFile string = "testEnv.json"

func TestMain(m *testing.M) {
	var (
		ecode int
		err   error
	)
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, r)
			ecode = 1
			fmt.Println(string(debug.Stack()))
		}
		if ecode != 0 {
			os.Exit(ecode)
		}
	}()

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
		panic(fmt.Errorf("error unmarshalling %v: %v\n",
			envFile, err.Error()))
	}

	pcc.InitDB(Env.DBConfiguration)   // Init the DB handler
	pcc.InitSSH(Env.SshConfiguration) // Init the SSH handler

	log.InitWithDefault(nil)

	credential := pcc.Credential{ // FIXME move to json
		UserName: "admin",
		Password: "admin",
	}
	if Pcc, err = pcc.Authenticate(Env.PccIp, credential); err != nil {
		panic(fmt.Errorf("Authentication error: %v\n", err))
	}

	dockerStats = pcc.InitDockerStats(Env.DockerStats)
	flag.Parse()
	if *test.DryRun {
		m.Run()
		return
	}

	ecode = m.Run()

	dockerStats.Stop()
	fmt.Println("\n\nTEST COMPLETED")
}

var count uint
var timeFormat = "Mon Jan 2 15:04:05 2006"

// TestNodes can be used to
// automatically config a cluser
func TestNodes(t *testing.T) {
	count++
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
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
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
	mayRun(t, "users", func(t *testing.T) {
		mayRun(t, "addTenant", addTenant)
	})
}

func TestMaaS(t *testing.T) {
	count++
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
	mayRun(t, "nodes", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "getSecKeys", getSecKeys)
		mayRun(t, "updateSecurityKey", updateSecurityKey_MaaS)
		mayRun(t, "addInvaders", addClusterHeads)
		mayRun(t, "addBrownfieldNodes", addBrownfieldServers)
		mayRun(t, "configServerInterfaces", configServerInterfaces)
		mayRun(t, "installLLDP", updateNodes_installLLDP)
		mayRun(t, "installMAAS", updateNodes_installMAAS)
		mayRun(t, "reimageAllBrownNodes", reimageAllBrownNodes)
	})
}

func TestTenantMaaS(t *testing.T) {
	count++
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
	mayRun(t, "nodes", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "getSecKeys", getSecKeys)
		mayRun(t, "updateSecurityKey", updateSecurityKey_MaaS)
		mayRun(t, "addInvaders", addClusterHeads)
		mayRun(t, "addBrownfieldNodes", addBrownfieldServers)
		mayRun(t, "configServerInterfaces", configServerInterfaces)
		mayRun(t, "installLLDP", updateNodes_installLLDP)
		mayRun(t, "installMAAS", updateNodes_installMAAS)
		mayRun(t, "addTenant", addTenant)
		mayRun(t, "reimageTenantAllBrownNodes", reimageAllBrownNodes)
	})
}

func TestAddK8s(t *testing.T) {
	count++
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
	mayRun(t, "nodes", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "addInvaders", addClusterHeads)
		mayRun(t, "addBrownfieldNodes", addBrownfieldServers)
		mayRun(t, "installLLDP", updateNodes_installLLDP)
		mayRun(t, "configServerInterfaces", configServerInterfaces)
		mayRun(t, "addIpam", updateIpam)
		mayRun(t, "addNetCluster", addNetCluster)
		mayRun(t, "CreateK8sCluster", createK8sCluster)
	})
}

func TestDeleteK8s(t *testing.T) {
	count++
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
	mayRun(t, "nodes", func(t *testing.T) {
		mayRun(t, "deleteK8sCluster", deleteK8sCluster)
		mayRun(t, "deleteNetCluster", deleteNetCluster)
	})
}

func TestNetCluster(t *testing.T) {
	count++
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
	mayRun(t, "netCluster", func(t *testing.T) {
		mayRun(t, "getNodesList", getNodes)
		mayRun(t, "addIpam", updateIpam)
		mayRun(t, "addNetCluster", addNetCluster)
	})
}

// assumes TestNode has been run before
func TestAddCeph(t *testing.T) {
	count++
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
	mayRun(t, "ceph", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "addIpam", updateIpam)
		mayRun(t, "addNetCluster", addNetCluster)
		mayRun(t, "testCeph", testCeph)
	})
}

func TestCephCache(t *testing.T) {
	mayRun(t, "ceph-cache", func(t *testing.T) {
		mayRun(t, "testCephCacheSetup", testCephCacheSetup)
		mayRun(t, "testCephCacheAdd", testCephCacheAdd)
		mayRun(t, "testCephCacheDelete", testCephCacheDelete)
	})
}

func TestK8sApp(t *testing.T) {
	count++
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
	mayRun(t, "K8sApp", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "addIpam", updateIpam)
		mayRun(t, "addNetCluster", addNetCluster)
		mayRun(t, "testK8sApp", testK8sApp)
	})
}

func TestPortus(t *testing.T) {
	count++
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
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

func TestDelPortus(t *testing.T) {
	count++
	fmt.Printf("Environment:\n%v\n", Env)
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
	mayRun(t, "portus", func(t *testing.T) {
		mayRun(t, "getAvailableNodes", getAvailableNodes)
		mayRun(t, "delAllPortus", delAllPortus)
	})
}

func TestHardwareInventory(t *testing.T) {
	count++
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
	mayRun(t, "hardwareinventory", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "addInvaders", addClusterHeads)
		mayRun(t, "installLLDP", updateNodes_installLLDP)
		mayRun(t, "installMAAS", updateNodes_installMAAS)
		mayRun(t, "testHardwareInventory", testHardwareInventory)
	})
}

func TestFull(t *testing.T) {
	count++
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
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
		// mayRun(t, "addSite", addSite)
		mayRun(t, "reimageTenantAllBrownNodes", reimageAllBrownNodes)
		mayRun(t, "CreateK8sCluster", createK8sCluster)
	})
}

func TestClean(t *testing.T) {
	count++
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
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
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
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
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
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
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
	mayRun(t, "AVAILABILITY", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "addInvaders", addClusterHeads)
		mayRun(t, "checkAddUnreachableNode", addUnreachableNode)
		mayRun(t, "checkAddInaccessibleNode", addInaccessibleNode)
		mayRun(t, "checkAgentAndCollectorRestore", checkAgentAndCollectorRestore)
	})
}

func TestGreenfield(t *testing.T) {
	count++
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
	mayRun(t, "GREENFIELD", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "updateSecurityKey", updateSecurityKey_MaaS)
		mayRun(t, "addInvaders", addClusterHeads)
		mayRun(t, "installLLDPOnInvaders", installLLDPOnInvaders)
		mayRun(t, "installMAAS", installMAAS)
		mayRun(t, "addGreenfieldNodes", addGreenfieldServers)
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "configNetworkInterfaces", configNetworkInterfaces)
		mayRun(t, "reimageAllBrownNodes", reimageAllBrownNodes)
	})
}

func TestConfigNetworkInterfaces(t *testing.T) {
	count++
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
	mayRun(t, "configureNetwork", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "configNetworkInterfaces", configNetworkInterfaces)
	})
}

func TestMonitor(t *testing.T) {
	count++
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
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
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
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
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
	mayRun(t, "KEY-MANAGER", func(t *testing.T) {
		mayRun(t, "testKMKeys", testKMKeys)
		mayRun(t, "testKMCertificates", testKMCertificates)
	})
}

func TestAppCredentials(t *testing.T) {
	count++
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
	mayRun(t, "APP CREDENTIALS", func(t *testing.T) {
		mayRun(t, "testCreateCredendialMetadataProfile", testCreateCredendialMetadataProfile)
		mayRun(t, "testUpdateCredendialMetadataProfile", testUpdateCredendialMetadataProfile)
		mayRun(t, "testDeleteCredendialMetadataProfile", testDeleteCredendialMetadataProfile)
	})
}

// Test functions for New Dashboard
func TestDashboard(t *testing.T) {
	count++
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
	mayRun(t, "DASHBOARD REST API", func(t *testing.T) {
		mayRun(t, "testDashboardGetAllPCCObjects", testDashboardGetAllPCCObjects)
		mayRun(t, "testDashboardGetPCCObjectByRandomId", testDashboardGetPCCObjectByRandomId)
		mayRun(t, "testDashboardGetPCCObjectById", testDashboardGetPCCObjectById)
		mayRun(t, "testDashboardGetChildrenObjectsByRandomId", testDashboardGetChildrenObjectsByRandomId)
		mayRun(t, "testDashboardGetParentObjectsByRandomId", testDashboardGetParentObjectsByRandomId)
		mayRun(t, "testDashboardGetFilteredObjects", testDashboardGetFilteredObjects)
		mayRun(t, "testDashboardGetAdvSearchedObjects", testDashboardGetAdvSearchedObjects)
		mayRun(t, "testDashboardGetAggrHealthCountByType", testDashboardGetAggrHealthCountByType)
		mayRun(t, "testDashboardGetMetadataEnumStrings", testDashboardGetMetadataEnumStrings)
	})
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
	fmt.Println("---")
	defer fmt.Println("...")
	fmt.Println("pcc instance unknown")
}
