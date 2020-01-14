package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
	"github.com/platinasystems/tiles/pccserver/models"
)

func reimageAllBrownNodes(t *testing.T) {
	t.Run("updateMAASInfo", updateMAASInfo)
	t.Run("reimageAllBrown", reimageAllBrown)
}

func updateMAASInfo(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var (
		body []byte
		resp HttpResp
	)
	for _, i := range Env.Servers {
		var (
			node models.NodeWithKubernetes
			data []byte
			err  error
		)
		keyId, err := getFirstKey()
		keys := pq.Int64Array{int64(keyId.Id)}

		pBool := new(bool)
		*pBool = true
		addReq := models.Node{
			Host:        i.HostIp,
			Id:          NodebyHostIP[i.HostIp],
			Bmc:         i.BMCIp,
			BmcUser:     i.BMCUser,
			BmcUsers:    i.BMCUsers,
			BmcPassword: i.BMCPass,
			AdminUser:   "admin",
			SSHKeys:     keys,
			Managed:     pBool,
			Console:     "ttyS1",
		}
		endpoint := fmt.Sprintf("pccserver/node/update")
		if data, err = json.Marshal(addReq); err != nil {
			assert.Fatalf("invalid struct for node update request")
		}
		resp, body, err = pccGateway("PUT", endpoint, data)
		if err != nil {
			assert.Fatalf("%v\n%v\n", string(body), err)
			return
		}
		if resp.Status != 200 {
			assert.Fatalf("%v\n", string(body))
			fmt.Printf("Update node %v failed\n%v\n",
				i.HostIp, string(body))
			return
		}
		if err := json.Unmarshal(resp.Data, &node); err != nil {
			assert.Fatalf("%v\n%v\n", string(resp.Data), err)
			return
		}
	}
}

func reimageAllBrown(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	key, err := getFirstKey()
	keys := []string{key.Alias}

	nodesList := make([]uint64, len(Env.Servers))
	for i, s := range Env.Servers {
		nodesList[i] = NodebyHostIP[s.HostIp]
	}

	var request pcc.MaasRequest
	request.Nodes = nodesList
	request.Image = "centos76"
	request.Locale = "en-US"
	request.Timezone = "PDT"
	request.AdminUser = "admin"
	request.SSHKeys = keys

	fmt.Println(request)
	if err = Pcc.MaasDeploy(request); err != nil {
		assert.Fatalf("MaasDeploy failed: %v\n", err)
	}

	fmt.Println("Sleep for 8 minutes")
	time.Sleep(8 * time.Minute)

	for {
		for i, id := range nodesList {
			status, err := getProvisionStatus(id)
			if err != nil {
				fmt.Printf("Node %v error: %v\n", id, err)
				continue
			}
			if strings.Contains(status, "Ready") {
				fmt.Printf("Node %v has gone Ready\n", id)
				nodesList = removeIndex(i, nodesList)
				continue
			} else if strings.Contains(status, "reimage failed") {
				fmt.Printf("Node %v has failed reimage\n", id)
				nodesList = removeIndex(i, nodesList)
				continue
			}
			fmt.Printf("Node %v: %v\n", id, status)
		}
		if len(nodesList) == 0 {
			fmt.Printf("Brownfield re-image done\n")
			return
		}
		time.Sleep(60 * time.Second)
	}
}

func removeIndex(i int, n []uint64) []uint64 {
	if len(n) > 1 {
		n = append(n[:i], n[i+1:]...)
		return n
	}
	return nil
}
