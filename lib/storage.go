// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"encoding/json"
	"fmt"

	"github.com/platinasystems/tiles/pccserver/models"
)

type StorageChildrenTO struct {
	models.StorageChildrenTO
}

func (p *PccClient) GetStorageNode(id uint64) (storage StorageChildrenTO, err error) {
	var (
		endpoint string
		resp     HttpResp
	)

	endpoint = fmt.Sprintf("pccserver/storage/node/" + fmt.Sprint(id) + "")
	if resp, _, err = p.pccGateway("GET", endpoint, nil); err != nil {
		return
	}
	if resp.Status != 200 {
		err = fmt.Errorf("%v", resp.Error)
		return
	}
	if err = json.Unmarshal(resp.Data, &storage); err != nil {
		return
	}
	return
}
