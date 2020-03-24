// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"github.com/platinasystems/tiles/pccserver/maas/models"
)

type MaasRequest struct {
	models.MaasRequest
}

func (p *PccClient) MaasDeploy(maasReq MaasRequest) (err error) {
	err = p.Post("maas/deployments", &maasReq, nil)
	return
}
