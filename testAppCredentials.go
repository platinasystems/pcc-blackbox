package main

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	log "github.com/platinasystems/go-common/logs"
	m "github.com/platinasystems/pcc-blackbox/models"
	"github.com/platinasystems/pcc-models/authentication"
	models "github.com/platinasystems/pcc-models/items"
)

////
// TEST APP Credentials
////

var metadataProfiles = make(map[uint64]*authentication.AuthProfile)

//
// get profile's parameters
//
func getServiceParameters(t *testing.T, servicetype string, mandatory bool, str bool, numeric bool) (parameters map[string]models.ItemDescription) {
	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	var (
		description authentication.AuthenticationServiceDescription
		err         error
	)

	parameters = make(map[string]models.ItemDescription)
	if description, err = Pcc.GetAppCredentialTemplate(servicetype); err == nil {
		for i := range description.Parameters {
			parameter := description.Parameters[i]

			if !parameter.Mandatory && mandatory {
				continue
			} else if parameter.Type == "string" && !str {
				continue
			} else if (strings.HasPrefix(parameter.Type, "uint") || strings.HasPrefix(parameter.Type, "int")) && !numeric {
				continue
			}

			parameters[parameter.Name] = parameter
		}
	} else {
		cleanAndFail(t, res, err)
	}
	return
}

//
// Create a profile
//
func createProfile(t *testing.T, servicetype string) (created authentication.AuthProfile) {
	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	var (
		err        error
		profile    map[string]interface{}
		value      interface{}
		parameters map[string]models.ItemDescription
	)

	profile = make(map[string]interface{})

	parameters = getServiceParameters(t, servicetype, false, true, true)
	for i := range parameters {
		parameter := parameters[i]

		if parameter.Type == "string" {
			value = "blackbox"

			if strings.Contains(parameter.Name, "mail") {
				value = fmt.Sprintf("%s@platiansystems.net", value) // TODO template
			}
		} else if strings.HasPrefix(parameter.Type, "uint") || strings.HasPrefix(parameter.Type, "int") {
			value = 10
		} else if parameter.Type == "bool" {
			value = false
		}

		profile[parameter.Name] = value
	}

	//TODO files

	appCredential := authentication.AuthProfile{Name: fmt.Sprintf("blackbox%s", servicetype),
		Type:    servicetype,
		Profile: profile,
		Active:  true}

	log.AuctaLogger.Infof("\ncreating the metadata profile", appCredential)
	created, err = Pcc.CreateAppCredentialProfile(&appCredential)
	log.AuctaLogger.Infof("\ncreated the metadata profile", created)

	if err != nil {
		log.AuctaLogger.Errorf("\nError creating the profile. Cleaning %v", err)
		cleanAndFail(t, res, err)
	}

	return
}

//
// create profiles
//
func testCreateCredendialMetadataProfile(t *testing.T) {
	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	var (
		types []string
		err   error
	)
	log.AuctaLogger.Info("\nAPP CREDENTIALS: creating the metadata profile")
	if types, err = Pcc.GetAppCredentialServices(); err == nil {
		log.AuctaLogger.Infof("Supported types are:", types)

		for i := range types {
			created := createProfile(t, types[i])
			metadataProfiles[created.Id] = &created
		}
	} else {
		log.AuctaLogger.Errorf("\nError creating the profile. Cleaning %v", err)
		cleanAndFail(t, res, err)
	}
}

//
// update created profiles
//
func testUpdateCredendialMetadataProfile(t *testing.T) {
	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	log.AuctaLogger.Info("\nAPP CREDENTIALS: updating the metadata profile")
	var (
		err            error
		parameters     map[string]models.ItemDescription
		updatedProfile authentication.AuthProfile
		newValue       interface{}
	)
	for _, profile := range metadataProfiles {

		updatedFields := make(map[string][]interface{})
		innerProfile := profile.Profile.(map[string]interface{})

		parameters = getServiceParameters(t, profile.Type, false, true, true)
		for i := range parameters {
			parameter := parameters[i]

			if parameter.Mandatory || strings.Contains(parameter.Name, "mail") { // Update only not mandatory fields
				continue
			}

			switch parameter.Type {
			case "string": // Add -update
				newValue = fmt.Sprintf("%v-update", innerProfile[parameter.Name])
			case "bool": // Not
				b, _ := strconv.ParseBool(fmt.Sprintf("%v", innerProfile[parameter.Name]))
				newValue = !b
			default: // +1
				i, _ := strconv.Atoi(fmt.Sprintf("%v", innerProfile[parameter.Name]))
				newValue = i + 1
			}
			updatedFields[parameter.Name] = make([]interface{}, 2)
			updatedFields[parameter.Name][0] = innerProfile[parameter.Name] // Old parameter
			updatedFields[parameter.Name][1] = newValue                     // New parameter

			innerProfile[parameter.Name] = newValue
			log.AuctaLogger.Info(fmt.Sprintf("changing the parameter %s from %v to %v", parameter.Name, updatedFields[parameter.Name][0], updatedFields[parameter.Name][1]))
		}

		profile.Profile = innerProfile

		log.AuctaLogger.Info("Updating the metadata profile", profile.Type, profile.Id)
		if updatedProfile, err = Pcc.UpdateAppCredentialProfile(profile); err == nil {
			log.AuctaLogger.Info("Updated the metadata profile", updatedProfile)
			// Check fields
			for k, v := range updatedFields {
				receivedValue := fmt.Sprintf("%v", innerProfile[k])

				if fmt.Sprintf("%v", v[1]) != receivedValue {
					log.AuctaLogger.Errorf("\nError updating the profile. Expected value for %s is %v. received %v. Cleaning", k, v[1], receivedValue)
					cleanAndFail(t, res, err)
				} else {
					fmt.Printf("\nThe parameter %s was correclty updated from %v to %v", k, v[0], receivedValue)
				}

			}
		} else {
			log.AuctaLogger.Errorf("\nError updating the profile. Cleaning %v", err)
			cleanAndFail(t, res, err)
		}
	}
}

//
// delete created profiles
//
func testDeleteCredendialMetadataProfile(t *testing.T) {
	res := m.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	log.AuctaLogger.Info("\nAPP CREDENTIALS: deleting the metadata profile")
	var err error
	for id, profile := range metadataProfiles {
		log.AuctaLogger.Infof("Deleting the profile %d %s\n", id, profile.Name)
		if _, e := Pcc.DeleteAppCredential(id); e != nil {
			err = e
		}
	}

	if err != nil {
		msg := err.Error()
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}
}

func cleanAndFail(t *testing.T, res *m.TestResult, err error) {
	testDeleteCredendialMetadataProfile(t)
	msg := err.Error()
	res.SetTestFailure(msg)
	log.AuctaLogger.Error(msg)
	t.FailNow()
}
