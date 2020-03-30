// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
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
	err = p.Post(PROFILE_ENDPOINT, &authProfile, nil)
	return
}

func (p *PccClient) GetAuthProfiles() (authProfiles []AuthenticationProfile, err error) {
	err = p.Get(PROFILE_ENDPOINT, &authProfiles)
	return
}

func (p *PccClient) GetAuthProfileById(id uint64) (authProfile AuthenticationProfile, err error) {
	endpoint := fmt.Sprintf("%v/%v", PROFILE_ENDPOINT, id)
	err = p.Get(endpoint, &authProfile)
	return

}

func (p *PccClient) GetAuthProfileByName(name string) (authProfile *AuthenticationProfile, err error) {
	var authProfiles []AuthenticationProfile
	if err = p.Get(PROFILE_ENDPOINT, &authProfiles); err == nil {
		for i := range authProfiles {
			if authProfiles[i].Name == name {
				authProfile = &authProfiles[i]
				return
			}
		}

	}
	return

}

func (p *PccClient) DelAuthProfile(id uint64) (err error) {
	endpoint := fmt.Sprintf("%v/%v", PROFILE_ENDPOINT, id)
	err = p.Delete(endpoint, nil, nil)
	return
}
