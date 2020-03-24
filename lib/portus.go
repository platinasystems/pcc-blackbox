// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
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
	err = p.Get(PORTUS_ENDPOINT, &portusConfigs)
	return
}

func (p *PccClient) GetPortusNodeById(id uint64) (portusConfig PortusConfiguration, err error) {
	endpoint := fmt.Sprintf("%v/%v", PORTUS_ENDPOINT, id)
	err = p.Get(endpoint, &portusConfig)
	return
}

func (p *PccClient) DelPortusNode(id uint64, removeStorage bool) (err error) {
	endpoint := fmt.Sprintf("%v/%v?removeStorage=%v", PORTUS_ENDPOINT, id,
		removeStorage)
	err = p.Delete(endpoint, nil)
	return
}

func (p *PccClient) InstallPortusNode(portusConfig PortusConfiguration) (err error) {
	err = p.Post(PORTUS_ENDPOINT, &portusConfig, nil)
	return
}
