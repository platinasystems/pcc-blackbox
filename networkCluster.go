package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
	"time"

	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

var netClusterName string = "mynetcluster"

func addNetCluster(t *testing.T) {
	test.SkipIfDryRun(t)

	for _, netCluster := range Env.NetCluster {
		netClusterId, err := Pcc.FindNetClusterId(netCluster.Name)
		netClusterName = netCluster.Name
		if err == nil {
			fmt.Printf("Network cluster [%v] already exists [%v]\n",
				netCluster.Name, netClusterId)
			continue
		}
		addNetClusterInternal(t, netCluster)
	}
}

func deleteNetCluster(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	for _, netCluster := range Env.NetCluster {
		_, err := Pcc.FindNetClusterId(netCluster.Name)
		if err != nil {
			assert.Fatalf("Network cluster [%v] not found: %v\n",
				netCluster.Name, err)
			continue
		}
		netClusterName = netCluster.Name
		break
	}
	deleteNetClusterInternal(t)
}

func addNetClusterInternal(t *testing.T, netCluster netCluster) {
	assert := test.Assert{t}

	var (
		reqCluster  pcc.NetworkClusterReq
		igwPolicy   string
		controlCIDR string
		dataCIDR    string
		err         error
	)

	reqCluster.Name = netCluster.Name

	igwPolicy = netCluster.IgwPolicy
	switch igwPolicy {
	case pcc.IGW_UPSTREAM:
	case pcc.IGW_DEFAULT:
	default:
		assert.Fatal("Invalid IgwPolicy")
		return
	}
	reqCluster.IgwPolicy = igwPolicy

	controlCIDR = netCluster.ControlCIDR
	if controlCIDR == "" {
		assert.Fatal("No ControlCIDR defined")
		return
	}
	controlCIDRObj, err := Pcc.FindSubnetObj(controlCIDR)
	if err != nil {
		assert.Fatalf("ControlCIDR IPAM not found [%v]: %v\n",
			controlCIDR, err)
		return
	}
	reqCluster.ControlCIDRId = controlCIDRObj.Id

	dataCIDR = netCluster.DataCIDR
	if controlCIDR == "" {
		assert.Fatal("No DataCIDR defined")
		return
	}
	dataCIDRObj, err := Pcc.FindSubnetObj(dataCIDR)
	if err != nil {
		assert.Fatalf("dataCIDR IPAM not found [%v]: %v\n",
			dataCIDR, err)
		return
	}
	reqCluster.DataCIDRId = dataCIDRObj.Id

	var digitCheck = regexp.MustCompile(`^[0-9]+$`)

	nodes := make([]pcc.NetNode, len(netCluster.Nodes))
	for i, n := range netCluster.Nodes {
		nodes[i].Id = NodebyHostIP[n.IpAddr]
		if digitCheck.MatchString(n.LocalAs) {
			val, err := json.Number(n.LocalAs).Int64()
			if err != nil {
				fmt.Printf("Atoi convert failed: %v\n", err)
				continue
			}
			val2 := uint64(val)
			nodes[i].LocalAs = pcc.SetAsn(val2)
		}
		peers := len(n.BgpNeighbors)
		if peers > 0 {
			nodes[i].BgpNeighbors = make([]pcc.BgpPeer, peers)
			for j, p := range n.BgpNeighbors {
				nodes[i].BgpNeighbors[j].NeighborIp =
					p.NeighborIp
				if !digitCheck.MatchString(p.RemoteAs) {
					fmt.Printf("Invalid RemoteAs [%v]\n",
						p.RemoteAs)
					continue
				}
				val, err := json.Number(p.RemoteAs).Int64()
				if err != nil {
					fmt.Printf("Atoi convert failed: %v\n",
						err)
					continue
				}
				val2 := uint64(val)
				nodes[i].BgpNeighbors[j].RemoteAs = pcc.SetAsn(val2)
			}
		}
	}
	reqCluster.Nodes = nodes

	err = Pcc.AddNetCluster(&reqCluster)
	if err != nil {
		assert.Fatalf("AddNetCluster failed: %v\n", err)
		return
	}

	time.Sleep(1 * time.Second)
	netClusterObj, err := Pcc.FindNetClusterName(reqCluster.Name)
	if err != nil {
		assert.Fatalf("Network cluster [%v]: %v\n",
			reqCluster.Name, err)
		return
	}
	netClusterId := netClusterObj.Id
	timeout := time.After(15 * time.Minute)
	tick := time.Tick(10 * time.Second)
	done := false
	for !done {
		select {
		case <-timeout:
			assert.Fatal("Timed out waiting for network cluster\n")
			return
		case <-tick:
			netClusterObj, err := Pcc.GetNetClusterId(netClusterId)
			if err != nil {
				assert.Fatalf("Network cluster [%v]: %v\n",
					reqCluster.Name, err)
				return
			}
			fmt.Printf("deploy status [%v] %v%% health [%v]\n",
				netClusterObj.DeployStatus,
				netClusterObj.ProgressPercentage,
				netClusterObj.Health)
			switch netClusterObj.DeployStatus {
			case pcc.NETWORK_DEPLOY_STATUS_COMPLETED:
				done = true
			case pcc.NETWORK_DEPLOY_STATUS_FAILED:
				done = true
				assert.Fatal("Network deploy failed\n")
			case pcc.NETWORK_DEPLOY_STATUS_UPDATE_FAILED:
				done = true
				assert.Fatal("Network deploy update failed\n")
			default:
			}
		}
	}
}

func deleteNetClusterInternal(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	netClusterId, err := Pcc.FindNetClusterId(netClusterName)
	if err != nil {
		assert.Fatalf("FindNetClusterId failed: %v\n", err)
		return
	}

	err = Pcc.DelNetClusterWait(netClusterId, false)
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
		err = Pcc.DelNetCluster(nC.Id, false)
		if err != nil {
			assert.Fatalf("DelNetCluster %v failed: %v\n",
				nC.Id, err)
			return
		}

		err = Pcc.DelNetClusterWait(nC.Id, false)
		if err != nil {
			assert.Fatalf("DelNetCluster %v failed: %v\n",
				nC.Id, err)
			return
		}
	}
}
