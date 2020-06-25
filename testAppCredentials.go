package main

import (
	"fmt"
	"github.com/platinasystems/pcc-models/authentication"
	"strings"
	"testing"
)

////
// TEST APP Credentials
////

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
		t.Fatal(err)
	}
	return
}

// Create a profile
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
		t.Fatal(err)
	}

	return
}

// Update a profile
func updateProfile(t *testing.T, profile *authentication.AuthProfile) (created authentication.AuthProfile) {
	var (
		err            error
		parameters     map[string]authentication.ItemDescription
		updatedProfile authentication.AuthProfile
	)

	updatedFields := make(map[string][]interface{})
	innerProfile := profile.Profile.(map[string]interface{})

	parameters = getServiceParameters(t, profile.Type, false, true, false)
	for i := range parameters {
		parameter := parameters[i]

		if parameter.Mandatory || parameter.Type != "string" || strings.Contains(parameter.Name, "mail") { // Update only not mandatory fields
			continue
		}

		newValue := fmt.Sprintf("%v-update", innerProfile[parameter.Name])
		updatedFields[parameter.Name] = make([]interface{}, 2)
		updatedFields[parameter.Name][0] = innerProfile[parameter.Name] // Old parameter
		updatedFields[parameter.Name][1] = newValue                     // New parameter

		fmt.Println(fmt.Sprintf("\nchanging the parameter from %v to %v", updatedFields[parameter.Name][0], updatedFields[parameter.Name][1]))
	}

	fmt.Println("updating the metadata profile", profile.Type, profile.Id)
	if updatedProfile, err = Pcc.UpdateAppCredentialProfile(profile); err == nil {
		fmt.Println("updated the metadata profile", updatedProfile.Type, updatedProfile.Id)
		// Check fields
		for k, v := range updatedFields {
			receivedValue := fmt.Sprintf("%v", innerProfile[k])

			if fmt.Sprintf("%v", v[1]) != receivedValue {
				t.Fatalf("Error updating the profile. Expected value for %s is %v. received %v", k, v[1], receivedValue)
			} else {
				fmt.Printf("\nThe parameter %s was correclty updated from %v to %v", k, v[0], receivedValue)
			}

		}

	} else {
		t.Fatal(err)
	}

	return
}

// add, get, update and delete a profile
func testCredendialMetadataProfile(t *testing.T) {
	var (
		types []string
		err   error
	)
	fmt.Println("\nAPP CREDENTIALS: creating the metadata profile")
	if types, err = Pcc.GetAppCredentialServices(); err == nil {
		fmt.Println("Supported types are:", types)

		for i := range types {
			created := createProfile(t, types[i])

			updateProfile(t, &created)

			defer func() {
				fmt.Printf("Deleting the profile %d %s\n", created.Id, created.Name)
				Pcc.DeleteAppCredential(created.Id)
			}()
		}
	}

	if err != nil {
		t.Fatal(err)
	}

}
