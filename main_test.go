// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

var Env testEnv
var Pcc pcc.PccClient

var Nodes = make(map[uint64]*pcc.NodeWithKubernetes)
var SecurityKeys = make(map[string]*pcc.SecurityKey)
var NodebyHostIP = make(map[string]uint64)

func TestMain(m *testing.M) {
	var (
		ecode  int
		output []byte
		err    error
	)
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, r)
			ecode = 1
		}
		if ecode != 0 {
			os.Exit(ecode)
		}
	}()

	output, err = exec.Command("cat", "testEnv.json").Output()
	if err != nil {
		panic(fmt.Errorf("no testEnv.json found"))
	}

	if err = json.Unmarshal(output, &Env); err != nil {
		panic(fmt.Errorf("error unmarshalling testEnv.json\n %v",
			err.Error()))
	}

	credential := pcc.Credential{
		UserName: "admin",
		Password: "admin",
	}
	Pcc, err = pcc.Authenticate(Env.PccIp, credential)
	if err != nil {
		panic(fmt.Errorf("%v\n", err))
	}

	flag.Parse()
	if *test.DryRun {
		m.Run()
		return
	}

	ecode = m.Run()
}

var count uint
var timeFormat = "Mon Jan 2 15:04:05 2006"

// TestNodes can be used to
// automatically config a cluser
func TestNodes(t *testing.T) {
	count++
	fmt.Printf("Environment:\n%v\n", Env)
	fmt.Printf("Iteration %v, %v\n",
		count, time.Now().Format(timeFormat))
	mayRun(t, "nodes", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "getSecKeys", getSecKeys)
		mayRun(t, "updateSecurityKey", updateSecurityKey_MaaS)
		mayRun(t, "addInvaders", addClusterHeads)
		mayRun(t, "addBrownfieldNodes", addBrownfieldServers)
		mayRun(t, "installLLDP", updateNodes_installLLDP)
		mayRun(t, "installMAAS", updateNodes_installMAAS)
		mayRun(t, "configServerInterfaces", configServerInterfaces)
		mayRun(t, "updateBmcInfo", updateBmcInfo)
	})
}

func TestMaaS(t *testing.T) {
	count++
	fmt.Printf("Environment:\n%v\n", Env)
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
	fmt.Printf("Environment:\n%v\n", Env)
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
		mayRun(t, "addSite", addSite)
		mayRun(t, "reimageTenantAllBrownNodes", reimageAllBrownNodes)
	})
}

func TestK8s(t *testing.T) {
	count++
	fmt.Printf("Environment:\n%v\n", Env)
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
	mayRun(t, "nodes", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "addInvaders", addClusterHeads)
		mayRun(t, "addBrownfieldNodes", addBrownfieldServers)
		mayRun(t, "installLLDP", updateNodes_installLLDP)
		mayRun(t, "configServerInterfaces", configServerInterfaces)
		mayRun(t, "CreateK8sCluster", createK8sCluster)
	})
}

func TestDeleteK8s(t *testing.T) {
	count++
	fmt.Printf("Environment:\n%v\n", Env)
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
	mayRun(t, "nodes", func(t *testing.T) {
		mayRun(t, "deleteK8sCluster", deleteK8sCluster)
	})
}

func TestPortus(t *testing.T) {
	count++
	fmt.Printf("Environment:\n%v\n", Env)
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format("Mon Jan 2 15:04:05 2006"))
	mayRun(t, "portus", func(t *testing.T) {
		mayRun(t, "getNodesList", getNodes)
		//mayRun(t, "addBrownfieldNodes", addBrownfieldServers)
		mayRun(t, "uploadSecurityAuthProfileCertificate", UploadSecurityAuthProfileCert)
		mayRun(t, "addProfile", AddAuthenticationProfile)
		mayRun(t, "uploadSecurityPortusKey", UploadSecurityPortusKey)
		mayRun(t, "uploadSecurityPortusCertificate", UploadSecurityPortusCert)
		mayRun(t, "installPortus", AddPortus)
		mayRun(t, "checkPortusInstallation", CheckPortusInstallation)
	})
}

func TestHardwareInventory(t *testing.T) {
	count++
	fmt.Printf("Environment:\n%v\n", Env)
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format("Mon Jan 2 15:04:05 2006"))
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
	fmt.Printf("Environment:\n%v\n", Env)
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
		mayRun(t, "addSite", addSite)
		mayRun(t, "reimageTenantAllBrownNodes", reimageAllBrownNodes)
		mayRun(t, "CreateK8sCluster", createK8sCluster)
	})
}

func TestClean(t *testing.T) {
	count++
	fmt.Printf("Environment:\n%v\n", Env)
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format(timeFormat))
	mayRun(t, "nodes", func(t *testing.T) {
		mayRun(t, "getAvailableNodes", getAvailableNodes)
		mayRun(t, "deleteK8sCluster", deleteK8sCluster)
		mayRun(t, "delAllPortus", delAllPortus)
		mayRun(t, "delAllNodes", delAllNodes)
		mayRun(t, "delAllUsers", delAllUsers)
		mayRun(t, "delAllTenants", delAllTenants)
		mayRun(t, "delAllKeys", delAllKeys)
		mayRun(t, "delAllProfiles", delAllProfiles)
		mayRun(t, "delAllCerts", delAllCerts)
	})
}

func TestGen(t *testing.T) {
	// Not a real testcase, but can be used to generate a
	// testEnv.json file from existing PCC setup.
	genEnv()
	os.Exit(0)
}

func mayRun(t *testing.T, name string, f func(*testing.T)) bool {
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
