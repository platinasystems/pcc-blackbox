// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"encoding/json"
	"fmt"

	"github.com/platinasystems/tiles/pccserver/models"
)

type Site struct {
	models.Site
}

func (p PccClient) AddSite(siteReq Site) (err error) {
	var data []byte

	endpoint := fmt.Sprintf("pccserver/site/add")
	if data, err = json.Marshal(siteReq); err != nil {
		return
	}
	_, _, err = p.pccGateway("POST", endpoint, data)
	if err != nil {
		return
	}
	return
}

func (p PccClient) DelSite(siteReq Site) (err error) {
	var data []byte

	val := []int64{siteReq.Id}

	endpoint := fmt.Sprintf("pccserver/site/delete")
	if data, err = json.Marshal(val); err != nil {
		return
	}
	_, _, err = p.pccGateway("POST", endpoint, data)
	if err != nil {
		return
	}
	return
}

func (p PccClient) GetSites() (sites []Site, err error) {
	var resp HttpResp

	endpoint := fmt.Sprintf("pccserver/site")
	resp, _, err = p.pccGateway("GET", endpoint, nil)
	if err != nil {
		return
	}
	err = json.Unmarshal(resp.Data, &sites)
	return
}

func (p PccClient) FindSite(name string) (site Site, err error) {
	var sites []Site

	sites, err = p.GetSites()
	if err != nil {
		return
	}
	for _, s := range sites {
		if s.Name == name {
			site = s
			return
		}
	}
	err = fmt.Errorf("unable to find site [%v]\n", site)
	return
}
