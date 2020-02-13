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
	PROFILE_ENDPOINT = "pccserver/profile"
)

type AuthenticationProfile struct {
	models.AuthenticationProfile
}

type LDAPConfiguration struct {
	models.LDAPConfiguration
}

func (p PccClient) AddAuthProfile(authProfile AuthenticationProfile) (err error) {
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
			authProfile.Name, err)
		return
	}
	return
}

func (p PccClient) GetAuthProfileByName(name string) (authProfile *AuthenticationProfile, err error) {

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
