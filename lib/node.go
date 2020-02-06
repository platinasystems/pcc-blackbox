// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"encoding/json"
	"fmt"

	avro "github.com/platinasystems/tiles/pccagent/models"
	"github.com/platinasystems/tiles/pccserver/models"
)

type NodeWithAdditionalFields struct {
	models.NodeWithAdditionalFields
}

type NodeDetail struct {
	NodeWithAdditionalFields
	//HardwareInventory *HardwareInventory
	//Apps              []ProvisionedApp
	Roles      []Role             `gorm:"many2many:node_roles"`
	SystemData *avro.Status       `json:"systemData"`
	Interfaces []*InterfaceDetail `json:"interfaces"`
}

func (p PccClient) GetNodes() (nodes []*NodeDetail, err error) {
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

	var nodes []*NodeDetail

	if nodes, err = p.GetNodes(); err != nil {
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
