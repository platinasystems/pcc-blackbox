package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
	"time"

	log "github.com/platinasystems/go-common/logs"
	"github.com/platinasystems/pcc-blackbox/models"

	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

var netClusterName string = "mynetcluster"

func addNetCluster(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now(), "addNetCluster")
	CheckDependencies(t, res, Env.CheckNetClusters)

	for _, netCluster := range Env.NetCluster {
		netClusterId, err := Pcc.FindNetClusterId(netCluster.Name)
		netClusterName = netCluster.Name
		if err == nil {
			log.AuctaLogger.Warnf("Network cluster [%v] already exists [%v]\n",
				netCluster.Name, netClusterId)
			continue
		}
		addNetClusterInternal(t, netCluster)
	}
}

func deleteNetCluster(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now(), "deleteNetCluster")
	CheckDependencies(t, res, Env.CheckNetClusters)

	assert := test.Assert{t}

	for _, netCluster := range Env.NetCluster {
		_, err := Pcc.FindNetClusterId(netCluster.Name)
		if err != nil {
			msg := fmt.Sprintf("Network cluster [%v] not found: %v\n",
				netCluster.Name, err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
			continue
		}
		netClusterName = netCluster.Name
		break
	}
	deleteNetClusterInternal(t)
}

func addNetClusterInternal(t *testing.T, netCluster netCluster) {
	assert := test.Assert{t}

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now(), "addNetClusterInternal")

	var (
		reqCluster    pcc.NetworkClusterReq
		netClusterObj *pcc.NetworkClusterUI
		igwPolicy     string
		controlCIDR   string
		dataCIDR      string
		err           error
	)

	reqCluster.Name = netCluster.Name

	igwPolicy = netCluster.IgwPolicy
	switch igwPolicy {
	case pcc.IGW_UPSTREAM:
	case pcc.IGW_DEFAULT:
	default:
		msg := "Invalid IgwPolicy"
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}
	reqCluster.IgwPolicy = igwPolicy

	controlCIDR = netCluster.ControlCIDR
	if controlCIDR == "" {
		msg := "No ControlCIDR defined"
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}
	controlCIDRObj, err := Pcc.FindSubnetObj(controlCIDR)
	if err != nil {
		msg := fmt.Sprintf("ControlCIDR IPAM not found [%v]: %v\n",
			controlCIDR, err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}
	reqCluster.ControlCIDRId = controlCIDRObj.Id

	dataCIDR = netCluster.DataCIDR
	if controlCIDR == "" {
		msg := "No DataCIDR defined"
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}
	dataCIDRObj, err := Pcc.FindSubnetObj(dataCIDR)
	if err != nil {

		msg := fmt.Sprintf("dataCIDR IPAM not found [%v]: %v\n",
			dataCIDR, err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
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
				log.AuctaLogger.Warnf("Atoi convert failed: %v\n", err)
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
					log.AuctaLogger.Warnf("Invalid RemoteAs [%v]\n",
						p.RemoteAs)
					continue
				}
				val, err := json.Number(p.RemoteAs).Int64()
				if err != nil {
					log.AuctaLogger.Warnf("Atoi convert failed: %v\n",
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
		msg := fmt.Sprintf("AddNetCluster failed: %v\n", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}

	time.Sleep(1 * time.Second)
	netClusterObj, err = Pcc.FindNetClusterName(reqCluster.Name)
	if err != nil {
		msg := fmt.Sprintf("Network cluster [%v]: %v\n",
			reqCluster.Name, err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()

		return
	}
	netClusterId := netClusterObj.Id
	timeout := time.After(15 * time.Minute)
	tick := time.Tick(10 * time.Second)
	done := false
	for !done {
		select {
		case <-timeout:
			msg := "Timed out waiting for network cluster\n"
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
			return
		case <-tick:
			netClusterObj, err = Pcc.GetNetClusterId(netClusterId)
			if err != nil {
				msg := fmt.Sprintf("Network cluster [%v]: %v\n",
					reqCluster.Name, err)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				assert.FailNow()
				return
			}
			log.AuctaLogger.Infof("deploy status [%v] %v%% health [%v]\n",
				netClusterObj.DeployStatus,
				netClusterObj.ProgressPercentage,
				netClusterObj.Health)

			switch netClusterObj.DeployStatus {
			case pcc.NETWORK_DEPLOY_STATUS_COMPLETED:
				done = true
			case pcc.NETWORK_DEPLOY_STATUS_FAILED:
				done = true
				msg := "Network deploy failed\n"
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				assert.FailNow()
			default:
			}
		}
	}

	if netClusterObj.Health != pcc.NETWORK_HEALTH_OK {
		msg := fmt.Sprintf("Network deploy %s\n", netClusterObj.Health)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}
}

func deleteNetClusterInternal(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now(), "deleteNetClusterInternal")
	assert := test.Assert{t}

	netClusterId, err := Pcc.FindNetClusterId(netClusterName)
	if err != nil {
		msg := fmt.Sprintf("FindNetClusterId failed: %v\n", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}

	err = Pcc.DelNetClusterWait(netClusterId, false)
	if err != nil {
		msg := fmt.Sprintf("DelNetCluster failed: %v\n", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}

	return
}

func delAllNetsCluster(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now(), "delAllNetsCluster")
	assert := test.Assert{t}

	netCluster, err := Pcc.GetNetCluster()
	if err != nil {
		msg := fmt.Sprintf("GetNetCluster failed: %v\n", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}

	for _, nC := range netCluster {
		log.AuctaLogger.Infof("delete net cluster [%v]\n", nC.Id)
		err = Pcc.DelNetCluster(nC.Id, false)
		if err != nil {
			msg := fmt.Sprintf("DelNetCluster %v failed: %v\n",
				nC.Id, err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
			return
		}

		err = Pcc.DelNetClusterWait(nC.Id, false)
		if err != nil {
			msg := fmt.Sprintf("DelNetCluster %v failed: %v\n",
				nC.Id, err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
			return
		}
	}
}
