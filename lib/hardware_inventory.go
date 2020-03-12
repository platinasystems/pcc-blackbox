// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"github.com/platinasystems/tiles/pccserver/models"
)

type HardwareInventory struct {
	models.HardwareInventory
}

func (p *PccClient) GetHardwareInventory() (hwInventory []HardwareInventory, err error) {
	err = p.Get("pccserver/hardware-inventory", &hwInventory)
	return
}
