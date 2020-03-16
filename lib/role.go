// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"encoding/json"
	"fmt"

	"github.com/platinasystems/tiles/pccserver/models"
)

const (
	ROLE_LLDP = "LLDP"
	ROLE_MAAS = "MaaS"
)

type Role struct {
	models.Role
}

func (p *PccClient) GetRoles() (roles []*Role, err error) {
	var resp HttpResp

	endpoint := fmt.Sprintf("pccserver/roles")
	if resp, _, err = p.pccGateway("GET", endpoint, nil); err != nil {
		return
	}
	if resp.Status != 200 {
		err = fmt.Errorf("%v", resp.Error)
		return
	}
	if err = json.Unmarshal(resp.Data, &roles); err != nil {
		return
	}
	return
}

func (p *PccClient) FindRoleId(role string) (id uint64, err error) {
	var (
		roles []*Role
	)

	roles, err = p.GetRoles()
	if err != nil {
		return
	}
	for _, r := range roles {
		if r.Name == role {
			id = r.ID
			return
		}
	}
	err = fmt.Errorf("Error: role [%v] not found", role)
	return
}
