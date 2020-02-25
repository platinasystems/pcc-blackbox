// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"encoding/json"
	"fmt"

	"github.com/platinasystems/tiles/pccserver/models"
)

type Notification struct {
	models.Notification
}

func (p *PccClient) GetNotifications() (notifications []Notification, err error) {
	var (
		resp     HttpResp
		endpoint string
	)

	page := 0
	limit := 50
	endpoint = fmt.Sprintf("pccserver/notifications/history"+
		"?page=%vlimit=%v", page, limit)
	resp, _, err = p.pccGateway("GET", endpoint, nil)
	if err != nil {
		return
	}
	err = json.Unmarshal(resp.Data, &notifications)
	if err != nil {
		return
	}
	return
}
