// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"encoding/json"
	"fmt"

	"github.com/platinasystems/tiles/pccserver/maas/models"
)

type MaasRequest struct {
	models.MaasRequest
}

func (p PccClient) MaasDeploy(maasReq MaasRequest) (err error) {
	var (
		data []byte
		resp HttpResp
	)

	endpoint := fmt.Sprintf("maas/deployments")
	if data, err = json.Marshal(maasReq); err != nil {
		return
	}

	if resp, _, err = p.pccGateway("POST", endpoint, data); err != nil {
		return
	}
	if resp.Status != 200 {
		err = fmt.Errorf("%v", resp.Error)
		return
	}
	return
}
