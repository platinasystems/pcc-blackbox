package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
	"time"

	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/pcc-blackbox/models"
	"github.com/platinasystems/test"
)

var netClusterName string = "mynetcluster"

func addNetCluster(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res, Env.CheckNetClusters, CheckNodesNetCluster)

	for _, netCluster := range Env.NetCluster {
		netClusterId, err := Pcc.FindNetClusterId(netCluster.Name)
		netClusterName = netCluster.Name
		if err == nil {
			log.AuctaLogger.Infof("Network cluster [%v] already "+
				"exists [%v]", netCluster.Name, netClusterId)
			health, summary, err :=
				Pcc.GetNetClusterHealthConn(netClusterId)
			if err != nil {
				msg := fmt.Sprintf("Failed to get net health "+
					"[%v]: %v", netClusterName, err)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				assert.FailNow()
			}
			if health != pcc.NETWORK_HEALTH_OK {
				msg := fmt.Sprintf("Net cluster health NotOK "+
					"[%v]: %v", netClusterName, summary)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				assert.FailNow()
			}
			log.AuctaLogger.Infof("Network Health [%v] [%v]",
				netClusterName, health)

			continue
		}
		addNetClusterInternal(t, netCluster)
	}
}

func deleteNetCluster(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res, Env.CheckNetClusters, CheckNetClusterExists)

	assert := test.Assert{t}

	for _, netCluster := range Env.NetCluster {
		_, err := Pcc.FindNetClusterId(netCluster.Name)
		if err != nil {
			msg := fmt.Sprintf("Network cluster [%v] not found: %v",
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
	defer res.CheckTestAndSave(t, time.Now())

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
	}
	reqCluster.IgwPolicy = igwPolicy

	controlCIDR = netCluster.ControlCIDR
	if controlCIDR == "" {
		msg := "No ControlCIDR defined"
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
	controlCIDRObj, err := Pcc.FindSubnetObj(controlCIDR)
	if err != nil {
		msg := fmt.Sprintf("ControlCIDR IPAM not found [%v]: %v",
			controlCIDR, err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
	reqCluster.ControlCIDRId = controlCIDRObj.Id

	dataCIDR = netCluster.DataCIDR
	if controlCIDR == "" {
		msg := "No DataCIDR defined"
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
	dataCIDRObj, err := Pcc.FindSubnetObj(dataCIDR)
	if err != nil {

		msg := fmt.Sprintf("dataCIDR IPAM not found [%v]: %v",
			dataCIDR, err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
	reqCluster.DataCIDRId = dataCIDRObj.Id

	var digitCheck = regexp.MustCompile(`^[0-9]+$`)

	nodes := make([]pcc.NetNode, len(netCluster.Nodes))
	for i, n := range netCluster.Nodes {
		nodes[i].Id = NodebyHostIP[n.IpAddr]
		if digitCheck.MatchString(n.LocalAs) {
			val, err := json.Number(n.LocalAs).Int64()
			if err != nil {
				log.AuctaLogger.Errorf("Atoi convert failed: "+
					"%v", err)
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
					log.AuctaLogger.Errorf("Invalid "+
						"RemoteAs [%v]", p.RemoteAs)
					continue
				}
				val, err := json.Number(p.RemoteAs).Int64()
				if err != nil {
					log.AuctaLogger.Errorf("Atoi convert "+
						"failed: %v", err)
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
		msg := fmt.Sprintf("AddNetCluster failed: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	time.Sleep(1 * time.Second)
	netClusterObj, err = Pcc.FindNetClusterName(reqCluster.Name)
	if err != nil {
		msg := fmt.Sprintf("Network cluster [%v]: %v",
			reqCluster.Name, err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
	netClusterId := netClusterObj.Id
	timeout := time.After(15 * time.Minute)
	tick := time.Tick(10 * time.Second)
	done := false
	notOkCount := 0
	for !done {
		select {
		case <-timeout:
			msg := "Timed out waiting for network cluster"
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
		case <-tick:
			netClusterObj, err = Pcc.GetNetClusterId(netClusterId)
			if err != nil {
				msg := fmt.Sprintf("Network cluster [%v]: %v",
					reqCluster.Name, err)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				assert.FailNow()
			}
			log.AuctaLogger.Infof("deploy status [%v] %v%% "+
				"health [%v]",
				netClusterObj.DeployStatus,
				netClusterObj.ProgressPercentage,
				netClusterObj.Health)

			switch netClusterObj.DeployStatus {
			case pcc.NETWORK_DEPLOY_STATUS_COMPLETED:
				switch netClusterObj.Health {
				case pcc.NETWORK_HEALTH_OK:
					done = true
				case pcc.NETWORK_HEALTH_NOT_OK:
					notOkCount++
					if notOkCount == 5 {
						done = true
					}
				}
			case pcc.NETWORK_DEPLOY_STATUS_FAILED:
				done = true
				msg := "Network deploy failed"
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				assert.FailNow()
			default:
			}
		}
	}

	// check connectivity
	health, summary, err := Pcc.GetNetClusterHealthConn(netClusterId)
	if err != nil {
		msg := fmt.Sprintf("Failed to get net health [%v]: %v",
			reqCluster.Name, err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
	log.AuctaLogger.Infof("Network Health & connectivity check [%v] [%v]",
		reqCluster.Name, health)
	if health != pcc.NETWORK_HEALTH_OK {
		msg := fmt.Sprintf("Net cluster health NotOK [%v]: %v",
			netClusterName, summary)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
}

func deleteNetClusterInternal(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	assert := test.Assert{t}

	netClusterId, err := Pcc.FindNetClusterId(netClusterName)
	if err != nil {
		msg := fmt.Sprintf("FindNetClusterId failed: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	err = Pcc.DelNetClusterWait(netClusterId, false)
	if err != nil {
		msg := fmt.Sprintf("DelNetCluster failed: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	return
}

func delAllNetsCluster(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	assert := test.Assert{t}

	netCluster, err := Pcc.GetNetCluster()
	if err != nil {
		msg := fmt.Sprintf("GetNetCluster failed: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	for _, nC := range netCluster {
		log.AuctaLogger.Infof("delete net cluster [%v]", nC.Id)
		err = Pcc.DelNetCluster(nC.Id, false)
		if err != nil {
			msg := fmt.Sprintf("DelNetCluster %v failed: %v",
				nC.Id, err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
		}

		err = Pcc.DelNetClusterWait(nC.Id, false)
		if err != nil {
			msg := fmt.Sprintf("DelNetCluster %v failed: %v",
				nC.Id, err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
		}
	}
}
