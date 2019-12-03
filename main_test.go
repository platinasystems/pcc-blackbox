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

	"github.com/platinasystems/test"
)

var Token string
var Bearer string
var Env testEnv

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
		panic(fmt.Errorf("error unmarshalling testEnv.json"))
	}

	type credential struct {
		UserName string `json:"username"`
		Password string `json:"password"`
	}

	postData, _ := json.Marshal(credential{UserName: "admin", Password: "admin"})
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

func Test(t *testing.T) {
	count++
	fmt.Printf("Environment:\n%v", Env)
	fmt.Printf("Iteration %v, %v\n", count, time.Now().Format("Mon Jan 2 15:04:05 2006"))
	mayRun(t, "nodes", func(t *testing.T) {
		mayRun(t, "getNodeList", getNodes)
		mayRun(t, "getAvailableNodes", getAvailableNodes)
		mayRun(t, "getSecKeys", getSecKeys)
		mayRun(t, "updateSecurityKey", updateSecurityKey_MaaS)
		mayRun(t, "addInvaders", addClusterHeads)
		mayRun(t, "addBrownfieldNodes", addBrownfieldServers)
		mayRun(t, "installLLDP", updateNodes_installLLDP)
		mayRun(t, "installMAAS", updateNodes_installMAAS)
		mayRun(t, "reimageAllBrownNodes", reimageAllBrownNodes)
		//mayRun(t,"configNetworkIntefaces", configNetworkIntefaces)
		//mayRun(t,"CreateK8sCluster", createK8s_3nodes)
		//mayRun(t, "delNodes", delNodes)

	})
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
