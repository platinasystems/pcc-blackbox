package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/platinasystems/test"
	"github.com/platinasystems/tiles/pccserver/models"
	"testing"
)

const (
	PROFILE_ENDPOINT   = "pccserver/profile"
	LDAP_CERT_FILENAME = "test_ldap_crt"
)

var CurrentAuthProfileName string

func AddAuthenticationProfile(t *testing.T) {
	t.Run("addAuthProfile", addAuthProfile)
}

func UploadSecurityAuthProfileCert(t *testing.T) {
	t.Run("uploadSecurityAuthProfileCert", uploadCertificate_AuthProfile)
}

func uploadCertificate_AuthProfile(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	err := CreateFileAndUpload(LDAP_CERT_FILENAME, LDAP_CERT, CERT)
	if err != nil {
		assert.Fatalf(err.Error())
	}
}

func addAuthProfile(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	var (
		authProfile models.AuthenticationProfile
		body        []byte
		resp        HttpResp
	)

	if Env.AuthenticationProfile.Name == "" {
		fmt.Printf("Authenticatiom Profile is not defined in the configuration file")
		return
	}
	authProfile = Env.AuthenticationProfile

	certificate, err := GetCertificate(LDAP_CERT_FILENAME)
	if err != nil {
		fmt.Printf("Get certificate %s failed\n%v\n", LDAP_CERT_FILENAME, err)
	} else {
		if authProfile.Type == "ldap" {
			var ldapConfiguration models.LDAPConfiguration
			decodeError := mapstructure.Decode(authProfile.Profile, &ldapConfiguration)
			if decodeError == nil {
				ldapConfiguration.CertificateId = &certificate.Id
				authProfile.Profile = ldapConfiguration
			}
		}
	}

	var label string
	for i := 1; ; i++ {
		label = fmt.Sprintf(authProfile.Name+"_%d", i)
		CurrentAuthProfileName = label
		if existingProfile, _ := GetAuthProfileByName(label); existingProfile == nil {
			break
		}
	}
	authProfile.Name = label

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

func GetAuthProfileByName(name string) (authProfile *models.AuthenticationProfile, err error) {

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
