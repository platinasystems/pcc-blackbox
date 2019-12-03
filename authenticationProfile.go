package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/platinasystems/test"
	"io/ioutil"
	"testing"
	"tiles/pccserver/models"
)

const PROFILE_ENDPOINT = "pccserver/profile"
const PROFILE_JSON = "authProfile.json"

func addAuthenticationProfile(t *testing.T) {
	t.Run("addAuthProfile", addAuthProfile)
}

func addAuthProfile(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	var (
		authProfile models.AuthenticationProfile
		body        []byte
		resp        HttpResp
	)

	if err := buildAuthProfile(&authProfile); err != nil {
		assert.Fatalf("Authentication Profile creation failed\n%v\n", err)
	}

	data, err := json.Marshal(authProfile)
	if err != nil {
		assert.Fatalf("invalid struct for add authentication profile request")
	}

	if resp, body, err = pccGateway("POST", PROFILE_ENDPOINT, data); err != nil {
		assert.Fatalf("%v\n%v\n", string(body), err)
		return
	}
	if resp.Status != 200 {
		assert.Fatalf("%v\n", string(body))
		fmt.Printf("add Authenticatiom Profile %v failed\n%v\n", authProfile.Name, string(body))
		return
	}
}


func getAuthProfileByName(name string)(authProfile *models.AuthenticationProfile, err error){

	var authProfiles [] models.AuthenticationProfile
	resp, body, err := pccGateway("GET", PROFILE_ENDPOINT, nil);

	if err == nil {
		if resp.Status == 200 {
			if err = json.Unmarshal(resp.Data, &authProfiles); err == nil {
				for i := range authProfiles {
					if authProfiles[i].Name == name {
						return &authProfiles[i], err
					}
				}
			}
		} else {
			err = errors.New(string(body))
		}
	}

	return nil, err
}

func buildAuthProfile(authProfile *models.AuthenticationProfile) (err error) {
	confFile, err := ioutil.ReadFile(PROFILE_JSON)
	json.Unmarshal(confFile, authProfile)
	return err
}
