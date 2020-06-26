package main

import (
	"fmt"
	"github.com/platinasystems/pcc-models/authentication"
	"strconv"
	"strings"
	"testing"
)

////
// TEST APP Credentials
////

var metadataProfiles = make(map[uint64]*authentication.AuthProfile)

//
// get profile's parameters
//
func getServiceParameters(t *testing.T, servicetype string, mandatory bool, str bool, numeric bool) (parameters map[string]authentication.ItemDescription) {
	var (
		description authentication.AuthenticationServiceDescription
		err         error
	)

	parameters = make(map[string]authentication.ItemDescription)
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
		cleanAndFail(t, err)
	}
	return
}

//
// Create a profile
//
func createProfile(t *testing.T, servicetype string) (created authentication.AuthProfile) {
	var (
		err        error
		profile    map[string]interface{}
		value      interface{}
		parameters map[string]authentication.ItemDescription
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

	fmt.Println("\ncreating the metadata profile", appCredential)
	created, err = Pcc.CreateAppCredentialProfile(&appCredential)
	fmt.Println("\ncreated the metadata profile", created)

	if err != nil {
		fmt.Printf("\nError creating the profile. Cleaning %v", err)
		cleanAndFail(t, err)
	}

	return
}

//
// create profiles
//
func testCreateCredendialMetadataProfile(t *testing.T) {
	var (
		types []string
		err   error
	)
	fmt.Println("\nAPP CREDENTIALS: creating the metadata profile")
	if types, err = Pcc.GetAppCredentialServices(); err == nil {
		fmt.Println("Supported types are:", types)

		for i := range types {
			created := createProfile(t, types[i])
			metadataProfiles[created.Id] = &created
		}
	} else {
		fmt.Printf("\nError creating the profile. Cleaning %v", err)
		cleanAndFail(t, err)
	}
}

//
// update created profiles
//
func testUpdateCredendialMetadataProfile(t *testing.T) {
	fmt.Println("\nAPP CREDENTIALS: updating the metadata profile")
	var (
		err            error
		parameters     map[string]authentication.ItemDescription
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
			fmt.Println(fmt.Sprintf("changing the parameter %s from %v to %v", parameter.Name, updatedFields[parameter.Name][0], updatedFields[parameter.Name][1]))
		}

		profile.Profile = innerProfile

		fmt.Println("Updating the metadata profile", profile.Type, profile.Id)
		if updatedProfile, err = Pcc.UpdateAppCredentialProfile(profile); err == nil {
			fmt.Println("Updated the metadata profile", updatedProfile)
			// Check fields
			for k, v := range updatedFields {
				receivedValue := fmt.Sprintf("%v", innerProfile[k])

				if fmt.Sprintf("%v", v[1]) != receivedValue {
					fmt.Printf("\nError updating the profile. Expected value for %s is %v. received %v. Cleaning", k, v[1], receivedValue)
					cleanAndFail(t, err)
				} else {
					fmt.Printf("\nThe parameter %s was correclty updated from %v to %v", k, v[0], receivedValue)
				}

			}
		} else {
			fmt.Printf("\nError updating the profile. Cleaning %v", err)
			cleanAndFail(t, err)
		}
	}
}

//
// delete created profiles
//
func testDeleteCredendialMetadataProfile(t *testing.T) {
	fmt.Println("\nAPP CREDENTIALS: deleting the metadata profile")
	var err error
	for id, profile := range metadataProfiles {
		fmt.Printf("Deleting the profile %d %s\n", id, profile.Name)
		if _, e := Pcc.DeleteAppCredential(id); e != nil {
			err = e
		}
	}

	if err != nil {
		t.Fatal(err)
	}
}

func cleanAndFail(t *testing.T, err error) {
	testDeleteCredendialMetadataProfile(t)
	t.Fatal(err)
}
