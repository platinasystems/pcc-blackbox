// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"encoding/json"
	"fmt"

	"github.com/platinasystems/tiles/pccserver/models"
)

type HardwareInventory struct {
	models.HardwareInventory
}

func (p PccClient) GetHardwareInventory() (hwInventory []HardwareInventory, err error) {
	var (
		endpoint string
		resp     HttpResp
	)

	endpoint = fmt.Sprintf("pccserver/hardware-inventory")
	if resp, _, err = p.pccGateway("GET", endpoint, nil); err != nil {
		return
	}
	if resp.Status != 200 {
		err = fmt.Errorf("%v", resp.Error)
		return
	}
	if err = json.Unmarshal(resp.Data, &hwInventory); err != nil {
		return
	}
	return
}
