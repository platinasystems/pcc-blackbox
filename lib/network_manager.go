// Copyright © 2020-2021 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"fmt"
	"strings"
	"time"

	"github.com/platinasystems/tiles/pccserver/models"
	"github.com/platinasystems/tiles/pccserver/network"
)

const (
	IGW_UPSTREAM = models.IgwUpstream
	IGW_DEFAULT  = models.IgwDefault

	NETWORK_DEPLOY_STATUS_COMPLETED    = network.NETWORK_DEPLOY_STATUS_COMPLETED
	NETWORK_DEPLOY_STATUS_FAILED       = network.NETWORK_DEPLOY_STATUS_FAILED
	NETWORK_DEPLOY_STATUS_PRE_CHECK    = network.NETWORK_DEPLOY_STATUS_PRE_CHECK
	NETWORK_DEPLOY_STATUS_INSTALLING   = network.NETWORK_DEPLOY_STATUS_INSTALLING
	NETWORK_DEPLOY_STATUS_UPDATING     = network.NETWORK_DEPLOY_STATUS_UPDATING
	NETWORK_DEPLOY_STATUS_UNINSTALLING = network.NETWORK_DEPLOY_STATUS_UNINSTALLING
	NETWORK_DEPLOY_STATUS_HEALTH_CHECK = network.NETWORK_DEPLOY_STATUS_HEALTH_CHECK
	NETWORK_DEPLOY_STATUS_REACH_CHECK  = network.NETWORK_DEPLOY_STATUS_REACH_CHECK

	NETWORK_HEALTH_OK          = network.NETWORK_HEALTH_OK
	NETWORK_HEALTH_WARN        = network.NETWORK_HEALTH_WARN
	NETWORK_HEALTH_NOT_OK      = network.NETWORK_HEALTH_NOT_OK
	NETWORK_HEALTH_IN_PROGRESS = network.NETWORK_HEALTH_IN_PROGRESS
	NETWORK_HEALTH_UNDEFINED   = network.NETWORK_HEALTH_UNDEFINED
)

type AsnType models.AsnType

type NetworkClusterReq struct {
	Id                 uint64    `json:"id"`
	Name               string    `json:"name"`
	ControlCIDRId      uint64    `json:"controlCIDRId"`
	DataCIDRId         uint64    `json:"dataCIDRId"`
	DeployStatus       string    `json:"deploy_status"`
	IgwPolicy          string    `json:"igwPolicy"`
	ProgressPercentage int8      `json:"progressPercentage"`
	Nodes              []NetNode `json:"nodes"`
}

type NetNode struct {
	Id           uint64    `json:"id"`
	LocalAs      AsnType   `json:"local_as"`
	BgpNeighbors []BgpPeer `json:"bgp_neighbors"`
}

type BgpPeer struct {
	NeighborIp string  `json:"neighbor_ip"`
	RemoteAs   AsnType `json:"remote_as"`
}

type NodeId struct {
	Id uint64
}

type NetworkClusterUI struct {
	models.NetworkClusterUI
}

type NetworkClusterNode struct {
	models.NetworkClusterNode
}

func SetAsn(newAsn uint64) (asn2 AsnType) {
	asn2 = AsnType(newAsn)
	return
}

func (p *PccClient) GetNetCluster() (netCluster []*NetworkClusterUI, err error) {
	err = p.Get("pccserver/network/cluster", &netCluster)
	return
}

func (p *PccClient) GetNetClusterId(id uint64) (netCluster *NetworkClusterUI, err error) {
	endpoint := fmt.Sprintf("pccserver/network/cluster/%d", id)
	err = p.Get(endpoint, &netCluster)
	return
}

func (p *PccClient) GetNetClusterHealth(id uint64) (health string, summary string, err error) {
	var netCluster *NetworkClusterUI

	endpoint := fmt.Sprintf("pccserver/network/cluster/health/%d", id)
	err = p.Get(endpoint, &netCluster)
	if err == nil {
		health = netCluster.Health
		summary = netCluster.HealthSummary
	}
	return
}

// same as GetNetClusterHealth, but also do a ping check
func (p *PccClient) GetNetClusterHealthConn(id uint64) (health string, summary string, err error) {
	var netCluster *NetworkClusterUI

	endpoint := fmt.Sprintf("pccserver/network/cluster/health/conn/%d", id)
	err = p.Get(endpoint, &netCluster)
	if err == nil {
		health = netCluster.Health
		summary = netCluster.HealthSummary
	}
	return
}

func (p *PccClient) FindNetClusterName(name string) (netCluster *NetworkClusterUI, err error) {
	var netClusters []*NetworkClusterUI

	netClusters, err = p.GetNetCluster()
	if err != nil {
		return
	}

	for _, net := range netClusters {
		if net.Name == name {
			netCluster = net
			return
		}
	}

	err = fmt.Errorf("Net Cluster [%v] not found\n", name)
	return
}

func (p *PccClient) FindNetClusterId(name string) (id uint64, err error) {
	var netClusters []*NetworkClusterUI

	netClusters, err = p.GetNetCluster()
	if err != nil {
		return
	}

	for _, net := range netClusters {
		if net.Name == name {
			id = net.Id
			return
		}
	}

	err = fmt.Errorf("Net Cluster [%v] not found\n", name)
	return
}

func (p *PccClient) AddNetCluster(netClusterReq *NetworkClusterReq) (err error) {
	err = p.Post("pccserver/network/cluster", netClusterReq, nil)
	return
}

func (p *PccClient) UpdateNetCluster(netClusterReq *NetworkClusterReq) (err error) {

	err = p.Put("pccserver/network/cluster", netClusterReq, nil)
	return
}

func (p *PccClient) DelNetCluster(id uint64, force bool) (err error) {
	type delNetCluster struct {
		forceRemove bool
	}
	endpoint := fmt.Sprintf("pccserver/network/cluster/%d", id)
	req := delNetCluster{forceRemove: force}
	err = p.Delete(endpoint, &req, nil)
	return
}

func (p *PccClient) DelNetClusterWait(id uint64, force bool) (err error) {
	p.DelNetCluster(id, force)
	timeout := time.After(10 * time.Minute)
	tick := time.Tick(1 * time.Second)
	done := false
	for !done {
		select {
		case <-timeout:
			err = fmt.Errorf("Timed out waiting for network cluster")
			return
		case <-tick:
			_, err = p.GetNetClusterId(id)
			if err != nil {
				if strings.Contains(err.Error(),
					"record not found") {
					err = nil
				}
				return
			}
		}
	}
	return
}
