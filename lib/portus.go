// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"encoding/json"
	"fmt"

	"github.com/platinasystems/tiles/pccserver/models"
)

const (
	PORTUS_ENDPOINT = "pccserver/portus"
)

type PortusConfiguration struct {
	models.PortusConfiguration
}

func (p PccClient) InstallPortusNode(portusConfig PortusConfiguration) (err error) {

	var (
		data []byte
		resp HttpResp
	)

	data, err = json.Marshal(portusConfig)
	if err != nil {
		return
	}

	resp, _, err = p.pccGateway("POST", PORTUS_ENDPOINT, data)
	if err != nil {
		return
	}
	if resp.Status != 200 {
		err = fmt.Errorf("%v", resp.Error)
		return
	}
	return
}
