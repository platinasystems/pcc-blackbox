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
	PROFILE_ENDPOINT = "pccserver/v1/profile"
)

// couldn't re-use models.AuthenticationProfile as
// it has a field of type interface{}
type AuthenticationProfile struct {
	ID      uint64            `json:"id"`
	Name    string            `json:"name" validate:"required" gorm:"name"`
	Type    string            `json:"type" validate:"required,eq=LDAP" gorm:"type"`
	Tenant  uint64            `json:"tenant" gorm:"tenant"`
	Profile LDAPConfiguration `json:"profile,omitempty" gorm:"-"`
}

type LDAPConfiguration struct {
	models.LDAPConfiguration
}

func (p *PccClient) AddAuthProfile(authProfile AuthenticationProfile) (err error) {
	var (
		data []byte
		resp HttpResp
	)

	data, err = json.Marshal(authProfile)
	if err != nil {
		return
	}

	resp, _, err = p.pccGateway("POST", PROFILE_ENDPOINT, data)
	if err != nil {
		return
	}
	if resp.Status != 200 {
		err = fmt.Errorf("add Authenticatiom Profile %v failed: %v\n",
			authProfile.Name, resp.Error)
		return
	}
	return
}

func (p *PccClient) GetAuthProfiles() (authProfiles []AuthenticationProfile, err error) {
	var (
		resp HttpResp
	)

	resp, _, err = p.pccGateway("GET", PROFILE_ENDPOINT, nil)
	if err != nil {
		return
	}
	if resp.Status == 200 {
		err = json.Unmarshal(resp.Data, &authProfiles)
		return
	}

	err = fmt.Errorf("%v", resp.Error)
	return
}

func (p *PccClient) GetAuthProfileById(id uint64) (authProfile AuthenticationProfile, err error) {
	var (
		endpoint string
		resp     HttpResp
	)

	endpoint = fmt.Sprintf("%v/%v", PROFILE_ENDPOINT, id)

	resp, _, err = p.pccGateway("GET", endpoint, nil)
	if err != nil {
		return
	}
	if resp.Status == 200 {
		err = json.Unmarshal(resp.Data, &authProfile)
		return
	}

	err = fmt.Errorf("%v", resp.Error)
	return

}

func (p *PccClient) GetAuthProfileByName(name string) (authProfile *AuthenticationProfile, err error) {

	var (
		resp         HttpResp
		authProfiles []AuthenticationProfile
	)
	resp, _, err = p.pccGateway("GET", PROFILE_ENDPOINT, nil)
	if err != nil {
		return
	}
	if resp.Status == 200 {
		if err = json.Unmarshal(resp.Data, &authProfiles); err == nil {
			for i := range authProfiles {
				if authProfiles[i].Name == name {
					authProfile = &authProfiles[i]
					return
				}
			}
		} else {
			return
		}
	}

	err = fmt.Errorf("%v", resp.Error)
	return

}

func (p *PccClient) DelAuthProfile(id uint64) (err error) {
	var (
		resp     HttpResp
		endpoint string
	)

	endpoint = fmt.Sprintf("%v/%v", PROFILE_ENDPOINT, id)

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
