// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"github.com/platinasystems/tiles/pccserver/models"
)

type Site struct {
	models.Site
}

func (p *PccClient) AddSite(siteReq Site) (err error) {
	err = p.Post("pccserver/site/add", &siteReq, nil)
	return
}

func (p *PccClient) DelSite(siteReq Site) (err error) {
	err = p.Post("pccserver/site/delete", &siteReq, nil)
	return
}

func (p *PccClient) GetSites() (sites []Site, err error) {
	err = p.Get("pccserver/site", &sites)
	return
}

func (p *PccClient) FindSite(name string) (site Site, err error) {
	var sites []Site

	if sites, err = p.GetSites(); err == nil {
		for _, s := range sites {
			if s.Name == name {
				site = s
				break
			}
		}
	}
	return
}
