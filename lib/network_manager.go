// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"fmt"

	"github.com/platinasystems/tiles/pccserver/models"
)

const (
	IGW_UPSTREAM = "upstream"
	IGW_DEFAULT  = "default"
)

type NetworkClusterReq struct {
	Id                 uint64   `json:"id"`
	Name               string   `json:"name"`
	ControlCIDR        string   `json:"controlCIDR"`
	DeployStatus       string   `json:"deploy_status"`
	IgwPolicy          string   `json:"igwPolicy"`
	ProgressPercentage int8     `json:"progressPercentage"`
	Nodes              []NodeId `json:"nodes"`
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

func (p *PccClient) GetNetCluster() (netCluster []*NetworkClusterUI, err error) {
	err = p.Get("pccserver/network/cluster", &netCluster)
	return
}

func (p *PccClient) GetNetClusterId(id uint64) (netCluster *NetworkClusterUI, err error) {
	endpoint := fmt.Sprintf("pccserver/network/cluster/%d", id)
	err = p.Get(endpoint, &netCluster)
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
	// This call to POST blocks.  Might want a non-blocking version
	// and use the GET to check progress/error/timeout
	err = p.Post("pccserver/network/cluster", netClusterReq, nil)
	return
}

func (p *PccClient) UpdateNetCluster(netClusterReq *NetworkClusterReq) (err error) {

	err = p.Put("pccserver/network/cluster", netClusterReq, nil)
	return
}

func (p *PccClient) DelNetCluster(id uint64) (err error) {
	endpoint := fmt.Sprintf("pccserver/network/cluster/%d", id)
	err = p.Delete(endpoint, nil, nil)
	return
}
