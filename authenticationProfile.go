package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/mitchellh/mapstructure"
	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/pcc-blackbox/models"
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

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	assert := test.Assert{t}
	err := CreateFileAndUpload(LDAP_CERT_FILENAME, LDAP_CERT, pcc.CERT, 0)
	if err != nil {
		msg := err.Error()
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
}

func addAuthProfile(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())
	CheckDependencies(t, res, Env.CheckAuthenticationProfile)

	assert := test.Assert{t}

	var (
		authProfile pcc.AuthenticationProfile
	)

	authProfile = Env.AuthenticationProfile

	exist, certificate, err := Pcc.FindCertificate(LDAP_CERT_FILENAME)
	if err != nil {
		msg := fmt.Sprintf("Get certificate %s failed%v",
			LDAP_CERT_FILENAME, err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	} else if exist {
		if authProfile.Type == "LDAP" {
			var ldapConfiguration pcc.LDAPConfiguration

			decodeError := mapstructure.Decode(authProfile.Profile,
				&ldapConfiguration)
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
		existingProfile, _ := Pcc.GetAuthProfileByName(label)
		if existingProfile == nil {
			break
		}
	}
	authProfile.Name = label

	err = Pcc.AddAuthProfile(authProfile)
	if err != nil {
		msg := fmt.Sprintf("Error: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}
}

func delAllProfiles(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	assert := test.Assert{t}

	var (
		authProfiles []pcc.AuthenticationProfile
		err          error
		id           uint64
	)

	authProfiles, err = Pcc.GetAuthProfiles()
	if err != nil {
		msg := fmt.Sprintf("Failed to get auth profiles: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	for _, aP := range authProfiles {
		id = aP.ID
		log.AuctaLogger.Infof("Deleting auth profile %v", aP.Name)
		err = Pcc.DelAuthProfile(id)
		if err != nil {
			msg := fmt.Sprintf("Failed to delete auth profile %v: %v",
				id, err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
		}
		// seems to be syncronous. API should document
	}
}
