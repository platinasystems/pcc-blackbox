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
	PORTUS_ENDPOINT = "pccserver/v1/portus"
)

type PortusConfiguration struct {
	models.PortusConfiguration
}

func (p *PccClient) GetPortusNodes() (portusConfigs []PortusConfiguration, err error) {
	var (
		data []byte
		resp HttpResp
	)

	resp, _, err = p.pccGateway("GET", PORTUS_ENDPOINT, data)
	if err != nil {
		return
	}
	if resp.Status != 200 {
		err = fmt.Errorf("%v", resp.Error)
		return
	}
	err = json.Unmarshal(resp.Data, &portusConfigs)
	if err != nil {
		return
	}
	return
}

func (p *PccClient) GetPortusNodeById(id uint64) (portusConfig PortusConfiguration, err error) {
	var (
		endpoint string
		data     []byte
		resp     HttpResp
	)

	endpoint = fmt.Sprintf("%v/%v", PORTUS_ENDPOINT, id)
	resp, _, err = p.pccGateway("GET", endpoint, data)
	if err != nil {
		return
	}
	if resp.Status != 200 {
		err = fmt.Errorf("%v", resp.Error)
		return
	}
	err = json.Unmarshal(resp.Data, &portusConfig)
	if err != nil {
		return
	}
	return
}

func (p *PccClient) DelPortusNode(id uint64, removeStorage bool) (err error) {
	var (
		endpoint string
		resp     HttpResp
	)

	endpoint = fmt.Sprintf("%v/%v?removeStorage=%v", PORTUS_ENDPOINT, id,
		removeStorage)
	resp, _, err = p.pccGateway("DELETE", endpoint, nil)
	if err != nil {
		return
	}
	if resp.Status != 200 {
		err = fmt.Errorf("%v", resp.Error)
		return
	}
	return
}

func (p *PccClient) InstallPortusNode(portusConfig PortusConfiguration) (err error) {

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
		err = fmt.Errorf("%v: %v", resp.Error, resp.Message)
		return
	}
	return
}
