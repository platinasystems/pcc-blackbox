package main

import (
	"fmt"
	"testing"

	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

var netClusterName = "mynetcluster"

func addNetCluster(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	var (
		reqCluster   pcc.NetworkClusterReq
		igwPolicy    string
		controlCIDR  string
		netClusterId uint64
		err          error
	)

	netClusterId, err = Pcc.FindNetClusterId(netClusterName)
	if err == nil {
		fmt.Printf("Network cluster [%v] already exists node [%d]\n",
			netClusterName, netClusterId)
		return
	}

	igwPolicy = Env.NetCluster.IgwPolicy
	controlCIDR = Env.NetCluster.ControlCIDR

	if igwPolicy == "" {
		assert.Fatal("No IgwPolicy defined")
		return
	} else {
		switch igwPolicy {
		case pcc.IGW_UPSTREAM:
		case pcc.IGW_DEFAULT:
		default:
			assert.Fatalf("Invalid IgwPolicy")
			return
		}
	}

	if controlCIDR == "" {
		assert.Fatal("No ControlCIDR defined")
		return
	}

	numNodes := len(Env.Invaders) + len(Env.Servers)
	nodes := make([]pcc.NodeId, numNodes)

	var j = 0
	for _, n := range Env.Invaders {
		nodes[j].Id = NodebyHostIP[n.HostIp]
		j++
	}
	for _, n := range Env.Servers {
		nodes[j].Id = NodebyHostIP[n.HostIp]
		j++
	}
	reqCluster.Nodes = nodes
	reqCluster.Name = netClusterName
	reqCluster.ControlCIDR = controlCIDR
	reqCluster.IgwPolicy = igwPolicy

	err = Pcc.AddNetCluster(&reqCluster)
	if err != nil {
		assert.Fatalf("AddNetCluster failed: %v\n", err)
		return
	}

	// 2nd request should fail
	err = Pcc.AddNetCluster(&reqCluster)
	assert.NonNil(err)

	netCluster, err := Pcc.GetNetCluster()
	if err != nil {
		assert.Fatalf("GetNetCluster failed: %v\n", err)
		return
	}

	var id uint64

	for _, nC := range netCluster {
		fmt.Printf("get cluster [%v]\n", nC)
		id = nC.Id
	}

	cluster, err := Pcc.GetNetClusterId(id)
	if err != nil {
		fmt.Printf("GetNetClusterId failed: %v\n", err)
		return
	}

	fmt.Printf("cluster [%v]\n", cluster)

	return
}

func deleteNetCluster(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	netClusterId, err := Pcc.FindNetClusterId(netClusterName)
	if err != nil {
		assert.Fatalf("FindNetClusterId failed: %v\n", err)
		return
	}

	err = Pcc.DelNetCluster(netClusterId)
	if err != nil {
		assert.Fatalf("DelNetCluster failed: %v\n", err)
		return
	}

	return
}

func delAllNetsCluster(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	netCluster, err := Pcc.GetNetCluster()
	if err != nil {
		assert.Fatalf("GetNetCluster failed: %v\n", err)
		return
	}

	for _, nC := range netCluster {
		fmt.Printf("delete net cluster [%v]\n", nC.Id)
		err = Pcc.DelNetCluster(nC.Id)
		if err != nil {
			assert.Fatalf("DelNetCluster %v failed: %v\n",
				nC.Id, err)
			return
		}
	}

}
