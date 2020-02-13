package main

import (
	"fmt"
	"testing"

	"github.com/mitchellh/mapstructure"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

const (
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
		authProfile pcc.AuthenticationProfile
	)

	if Env.AuthenticationProfile.Name == "" {
		fmt.Printf("Authenticatiom Profile is not defined in the configuration file")
		return
	}
	authProfile = Env.AuthenticationProfile

	certificate, err := Pcc.FindCertificate(LDAP_CERT_FILENAME)
	if err != nil {
		fmt.Printf("Get certificate %s failed\n%v\n", LDAP_CERT_FILENAME, err)
	} else {
		if authProfile.Type == "ldap" {
			var ldapConfiguration pcc.LDAPConfiguration
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

	err = Pcc.AddAuthProfile(authProfile)
	if err != nil {
		assert.Fatalf("Error: %v\n", err)
		return
	}
}

func GetAuthProfileByName(name string) (authProfile *pcc.AuthenticationProfile, err error) {

	authProfile, err = Pcc.GetAuthProfileByName(name)
	return
}
