// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"encoding/json"
	"fmt"

	"github.com/platinasystems/tiles/pccserver/models"
)

type NodeAvailability struct {
	models.NodeAvailability
}

type Node struct {
	models.Node
}

type NodeWithAdditionalFields struct {
	models.NodeWithAdditionalFields
}

type NodeWithKubernetes struct {
	models.NodeWithAdditionalFields
}

type NodeDetail struct {
	//models.NodeDetail
	NodeWithAdditionalFields
	//HardwareInventory *HardwareInventory
	//Apps              []ProvisionedApp
	//Roles             []Role             `gorm:"many2many:node_roles"`
	// SystemData        *avro.Status       `json:"systemData"`
	Interfaces []*InterfaceDetail `json:"interfaces"`
}

func (p PccClient) GetNodesDetail() (nodes []*NodeDetail, err error) {
	var resp HttpResp

	endpoint := fmt.Sprintf("pccserver/node")
	if resp, _, err = p.pccGateway("GET", endpoint, nil); err != nil {
		return
	}
	if resp.Status != 200 {
		err = fmt.Errorf("%v", resp.Error)
		return
	}
	if err = json.Unmarshal(resp.Data, &nodes); err != nil {
		return
	}
	return
}

func (p PccClient) GetNodesWithKubernetes() (nodes []*NodeWithKubernetes,
	err error) {

	var resp HttpResp

	endpoint := fmt.Sprintf("pccserver/node")
	if resp, _, err = p.pccGateway("GET", endpoint, nil); err != nil {
		return
	}
	if resp.Status != 200 {
		err = fmt.Errorf("%v", resp.Error)
		return
	}
	if err = json.Unmarshal(resp.Data, &nodes); err != nil {
		return
	}
	return
}

func (p PccClient) GetNodesId(id uint64) (node NodeDetail, err error) {
	var resp HttpResp

	endpoint := fmt.Sprintf("pccserver/node/%v", id)
	if resp, _, err = p.pccGateway("GET", endpoint, nil); err != nil {
		return
	}
	if resp.Status != 200 {
		err = fmt.Errorf("%v", resp.Error)
		return
	}
	if err = json.Unmarshal(resp.Data, &node); err != nil {
		return
	}
	return
}

func (p PccClient) FindNodeId(name string) (id uint64, err error) {

	var nodes []*NodeWithKubernetes

	if nodes, err = p.GetNodesWithKubernetes(); err != nil {
		return
	}

	for _, n := range nodes {
		if n.Name == name {
			id = n.Id
			return
		}
	}
	err = fmt.Errorf("node [%v] not found", name)
	return
}

func (p PccClient) GetProvisionStatus(id uint64) (status string, err error) {

	var resp HttpResp

	endpoint := fmt.Sprintf("pccserver/node/%v/provisionStatus", id)
	if resp, _, err = p.pccGateway("GET", endpoint, nil); err != nil {
		return
	}
	if resp.Status == 200 {
		status = string(resp.Data) // status has double quotes
		return
	}
	err = fmt.Errorf("%v", resp.Error)
	return
}

func (p PccClient) GetNodeSummary(id uint64, node *NodeWithKubernetes) (err error) {

	var resp HttpResp

	endpoint := fmt.Sprintf("pccserver/node/summary/%v", id)
	if resp, _, err = p.pccGateway("GET", endpoint, nil); err != nil {
		return
	}
	if resp.Status == 200 {
		err = json.Unmarshal(resp.Data, node)
		if err != nil {
			return
		}
		return
	}
	err = fmt.Errorf("%v", resp.Message)
	return
}

func (p PccClient) GetNodeConnectionStatus(node *NodeWithKubernetes) (status string, err error) {
	if node.NodeAvailabilityStatus != nil {
		status = node.NodeAvailabilityStatus.ConnectionStatus
		return
	}
	err = fmt.Errorf("No NodeAvailablityStatus\n")
	return
}

func (p PccClient) AddNode(hostIp string, pccManaged bool) (node NodeWithKubernetes, err error) {

	var (
		addReq   NodeWithKubernetes
		data     []byte
		endpoint string
		resp     HttpResp
	)

	endpoint = fmt.Sprintf("pccserver/node/add/")
	addReq.Host = hostIp
	pBool := new(bool)
	*pBool = pccManaged
	addReq.Managed = pBool
	if data, err = json.Marshal(addReq); err != nil {
		return
	}

	resp, _, err = p.pccGateway("POST", endpoint, data)
	if err != nil {
		return
	}
	if resp.Status == 200 {
		err = json.Unmarshal(resp.Data, &node)
		if err != nil {
			return
		}
		return
	}
	err = fmt.Errorf("%v", resp.Message)
	return
}

func (p PccClient) UpdateNode(addReq NodeWithKubernetes) (node NodeWithKubernetes, err error) {

	var (
		data     []byte
		endpoint string
		resp     HttpResp
	)

	endpoint = fmt.Sprintf("pccserver/node/update")
	if data, err = json.Marshal(addReq); err != nil {
		return
	}

	resp, _, err = p.pccGateway("PUT", endpoint, data)
	if err != nil {
		return
	}
	if resp.Status == 200 {
		err = json.Unmarshal(resp.Data, &node)
		if err != nil {
			return
		}
		return
	}
	err = fmt.Errorf("%v", resp.Message)
	return
}

func (p PccClient) DelNode(id uint64) (err error) {
	var (
		endpoint string
		resp     HttpResp
	)

	endpoint = fmt.Sprintf("pccserver/node/%v", id)
	resp, _, err = p.pccGateway("DELETE", endpoint, nil)
	if err != nil {
		return
	}
	if resp.Status == 200 {
		return
	}
	err = fmt.Errorf("%v", resp.Message)
	return
}
