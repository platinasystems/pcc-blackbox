// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"fmt"
	"github.com/platinasystems/tiles/pccserver/models"
)

const ( // FIXME unify role and notification type
	ROLE_LLDP     = "LLDP"    // Phuket
	ROLE_DEFAULT  = "Default" // Takayama
	ROLE_LLDPD    = "LLDPD"
	ROLE_MAAS     = "MaaS"                      // Phuket
	ROLE_MAAS_NEW = "Baremetal Management Node" // Takayama
)

type Role struct {
	models.Role
}

func (p *PccClient) GetRoles() (roles []*Role, err error) {
	endpoint := fmt.Sprintf("pccserver/roles")
	err = p.Get(endpoint, &roles)
	return
}

// Find a role between more alternatives
func (p *PccClient) FindRoleId(role ...string) (id uint64, err error) {
	var (
		roles []*Role
	)

	if roles, err = p.GetRoles(); err == nil {
		for _, r := range roles {
			for _, desiredRole := range role {
				if r.Name == desiredRole {
					id = r.ID
					return
				}
			}
		}
		err = fmt.Errorf("error: role [%v] not found", role)
	}
	return
}
