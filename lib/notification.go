// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"fmt"

	"github.com/platinasystems/tiles/pccserver/models"
)

type Notification struct {
	models.Notification
}

func (p *PccClient) GetNotifications() (notifications []Notification, err error) {
	page := 0
	limit := 50
	endpoint := fmt.Sprintf("pccserver/notifications/history?page=%vlimit=%v", page, limit)
	err = p.Get(endpoint, &notifications)
	return
}
