// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"encoding/json"
	"fmt"

	"github.com/platinasystems/tiles/pccserver/models"
)

type ProvisionedApp struct {
	models.ProvisionedApp
}

func (p PccClient) GetApps(nodeId uint64) (apps []ProvisionedApp, err error) {

	var (
		endpoint string
		body     []byte
		resp     HttpResp
	)

	endpoint = fmt.Sprintf("pccserver/node/%v/apps", nodeId)
	if resp, _, err = p.pccGateway("GET", endpoint, nil); err != nil {
		return
	}
	if resp.Status != 200 {
		err = fmt.Errorf("Get node apps failed\n%v\n", string(body))
		return
	}
	if err = json.Unmarshal(resp.Data, &apps); err != nil {
		return
	}

	return
}

func (p PccClient) IsAppInstalled(nodeId uint64, appId string) (isInstalled bool) {
	var (
		apps []ProvisionedApp
		err  error
	)

	if apps, err = p.GetApps(nodeId); err != nil {
		fmt.Printf("Failed to GetApps: %v\n", err)
		return
	}

	for i := range apps {
		if apps[i].ID == appId && apps[i].Local.Installed {
			isInstalled = true
			return
		}
	}
	return
}
