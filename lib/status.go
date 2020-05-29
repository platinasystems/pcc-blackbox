// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"fmt"

	"github.com/platinasystems/tiles/pccserver/models"
)

const (
	OK             = models.OK   // done
	KO             = models.KO   // blocking error
	WARN           = models.WARN // non-blocking error
	RUNNING string = "running"
)

type Status struct {
	models.Status
}

func (p *PccClient) GetStatusId(id uint64, whichType string) (statuses []*Status, err error) {
	endpoint := fmt.Sprintf("pccserver/node/%v/status/%v", id, whichType)
	err = p.Get(endpoint, &statuses)
	return
}
