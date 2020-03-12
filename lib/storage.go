// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"fmt"
	"github.com/platinasystems/tiles/pccserver/models"
)

type StorageChildrenTO struct {
	models.StorageChildrenTO
}

func (p *PccClient) GetStorageNode(id uint64) (storage StorageChildrenTO, err error) {
	endpoint := fmt.Sprintf("pccserver/storage/node/" + fmt.Sprint(id) + "")
	err = p.Get(endpoint, &storage)
	return
}
