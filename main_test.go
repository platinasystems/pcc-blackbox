// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/platinasystems/tiles/pccserver/models"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

var Token string
var Bearer string
var Env testEnv

// redundant global till we migrate to golang binding
var Pcc pcc.PccClient

var Nodes = make(map[uint64]*models.NodeWithKubernetes)
var SecurityKeys = make(map[string]*securityKey)
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
			//test.Pause()
			os.Exit(ecode)
		}
	}()
	if output, err = exec.Command("cat", "testEnv.json").Output(); err != nil {
		panic(fmt.Errorf("no testEnv.json found"))
	}
	if err = json.Unmarshal(output, &Env); err != nil {
		panic(fmt.Errorf("error unmarshalling testEnv.json\n %v", err.Error()))
	}

	credential := pcc.Credential{
		UserName: "admin",
		Password: "admin",
	}

	Pcc, err = pcc.Authenticate(Env.PccIp, credential)
	if err != nil {
		panic(fmt.Errorf("%v\n", err))
	}

	postData, _ := json.Marshal(credential)
	url := fmt.Sprintf("https://%s:9999/security/auth", Env.PccIp)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	if resp, err := http.Post(url, "application/json", bytes.NewBuffer(postData)); err == nil {
		defer resp.Body.Close()
		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			var out struct{ Token string }
			json.Unmarshal(body, &out)
			Token = out.Token
			Bearer = "Bearer " + Token
			fmt.Printf("token: %v\n", Token)
		} else {
			fmt.Printf("Error getting token from pcc:\n%v\n", string(body))
			return
		}
	} else {
		fmt.Printf("err: %v\n", err)
	}

	flag.Parse()
	if *test.DryRun {
		m.Run()
		return
	}
	if testing.Verbose() {
		//uutInfo()
	}
	ecode = m.Run()
}

var count uint

// TestNodes can be used to
// automatically config a cluser
func TestNodes(t *testing.T) {
	count++
	fmt.Printf("Environment:\n%v\n", Env)
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format("Mon Jan 2 15:04:05 2006"))
	mayRun(t, "nodes", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "getSecKeys", getSecKeys)
		mayRun(t, "updateSecurityKey", updateSecurityKey_MaaS)
		mayRun(t, "addInvaders", addClusterHeads)
		mayRun(t, "addBrownfieldNodes", addBrownfieldServers)
		mayRun(t, "installLLDP", updateNodes_installLLDP)
		mayRun(t, "installMAAS", updateNodes_installMAAS)
		mayRun(t, "configServerInterfaces", configServerInterfaces)
		mayRun(t, "updateMAASInfo", updateMAASInfo)
	})
}

func TestMaaS(t *testing.T) {
	count++
	fmt.Printf("Environment:\n%v\n", Env)
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format("Mon Jan 2 15:04:05 2006"))
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
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format("Mon Jan 2 15:04:05 2006"))
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
		mayRun(t, "reimageAllBrownNodes", reimageAllBrownNodes)
	})
}

func TestK8s(t *testing.T) {
	count++
	fmt.Printf("Environment:\n%v\n", Env)
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format("Mon Jan 2 15:04:05 2006"))
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
	fmt.Printf("Iteration %v, %v\n", count,
		time.Now().Format("Mon Jan 2 15:04:05 2006"))
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
func TestClean(t *testing.T) {
	getAvailableNodes(t)
	delAllNodes(t)
	os.Exit(0)
}

func TestGen(t *testing.T) {
	// Not a real testcase, but can be used to generate a
	// testEnv.json file from existing PCC setup.
	getAvailableNodes(t)
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
