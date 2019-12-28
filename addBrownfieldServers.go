package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/platinasystems/test"
	"github.com/platinasystems/tiles/pccserver/models"
)

func addBrownfieldServers(t *testing.T) {
	t.Run("addBrownfieldServers", addServer)
}

func addServer(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var (
		body  []byte
		resp  HttpResp
		err   error
		check bool
	)
	nodesToCheck := make([]uint64, 0)
	for _, i := range Env.Servers {
		if Nodes[NodebyHostIP[i.HostIp]] != nil {
			continue
		}
		var (
			node models.NodeWithKubernetes
			data []byte
		)
		addReq := nodeAddReq{
			Host:    i.HostIp,
			Managed: true,
		}
		endpoint := fmt.Sprintf("pccserver/node/add")
		if data, err = json.Marshal(addReq); err != nil {
			assert.Fatalf("invalid struct for node add request")
		}
		if resp, body, err = pccGateway("POST", endpoint, data); err != nil {
			assert.Fatalf("%v\n%v\n", string(body), err)
			return
		}
		if resp.Status != 200 {
			assert.Fatalf("%v\n", string(body))
			fmt.Printf("add node %v failed\n%v\n", i.HostIp, string(body))
			return
		}
		if err := json.Unmarshal(resp.Data, &node); err != nil {
			assert.Fatalf("%v\n%v\n", string(resp.Data), err)
			return
		}
		if node.Id != 0 {
			node.Invader = false
			nodesToCheck = append(nodesToCheck, node.Id)
			Nodes[node.Id] = &node
			fmt.Printf("Add id %v to Nodes\n", node.Id)
			NodebyHostIP[node.Host] = node.Id
			fmt.Printf("Mapping hostIP %v to id %v\n", node.Host, node.Id)
		}
	}

	// early check for add fail
	time.Sleep(10 * time.Second)
	for id := range Nodes {
		if status, err := getProvisionStatus(id); err == nil {
			if strings.Contains(status, "Add node failed") {
				assert.Fatalf("%v for %v\n", status, id)
			}
		}
	}

	from := time.Now()
	//Check Agent installation
	//SERIAL - to be improved
	for i := 0; i < len(nodesToCheck); i++ {
		check = false
		fmt.Printf("Checking Agent installation for nodeId:%v\n", nodesToCheck[i])
		//check, err = checkAgentInstallation(nodesToCheck[i])
		check, err = checkGenericInstallation(nodesToCheck[i], AGENT_TIMEOUT, AGENT_NOTIFICATION, from)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
		if check {
			fmt.Printf("AGENT correctly installed on nodeId:%v\n", nodesToCheck[i])
		}
	}

	from = time.Now()
	//Check Collector installation
	for i := 0; i < len(nodesToCheck); i++ {
		fmt.Printf("Checking Collector installation for nodeId:%v\n", nodesToCheck[i])
		//check, err = checkCollectorInstallation(nodesToCheck[i])
		check, err = checkGenericInstallation(nodesToCheck[i], COLLECTOR_TIMEOUT, COLLECTOR_NOTIFICATION, from)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
		if check {
			fmt.Printf("COLLECTOR correctly installed on nodeId:%v\n", nodesToCheck[i])
		}
	}

	// wait for agent/collector to install
	time.Sleep(10 * time.Second)
	start := time.Now()
	done := false
	timeout := 90 * time.Second
	for !done {
		done = true
		for id, node := range Nodes {
			if node.NodeAvailabilityStatus != nil {
				if node.NodeAvailabilityStatus.ConnectionStatus == "online" {
					continue
				}
			}
			done = false
			err = getNodeSummary(id, node)
			if err != nil {
				fmt.Printf("node %v, error: %v\n", id, err)
				continue
			}
			name := fmt.Sprintf("node:%v", id)
			if node.Name != "" {
				name = node.Name
			}
			fmt.Printf("%v is %v provisionStatus = %v \n", name,
				node.NodeAvailabilityStatus.ConnectionStatus,
				node.ProvisionStatus)
			if node.ProvisionStatus == "Add node failed" {
				assert.Fatalf("%v for %v\n",
					node.ProvisionStatus, name)
			}
			done = node.NodeAvailabilityStatus.ConnectionStatus == "online"
			Nodes[id] = node
		}

		if !done {
			time.Sleep(10 * time.Second)
		}
		if time.Since(start) > timeout {
			break
		}
	}

	if !done {
		for _, node := range Nodes {
			if node.NodeAvailabilityStatus == nil {
				assert.Fatalf("node %v did not come online; provisionStatus = %v\n", node.Name, node.ProvisionStatus)
			} else {
				if node.NodeAvailabilityStatus.ConnectionStatus != "online" {
					assert.Fatalf("node %v did not come online; provisionStatus = %v\n", node.Name, node.ProvisionStatus)
				}
			}
		}
	}
}
