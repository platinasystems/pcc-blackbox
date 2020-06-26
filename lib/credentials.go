package pcc

import (
	"encoding/json"
	"fmt"
	"github.com/platinasystems/pcc-models/authentication"
)

////
// Credentials Rest Client
////

// Get app credentials
func (pcc *PccClient) GetAppCredentials(serviceType string) (profiles []authentication.AuthProfile, err error) {
	err = pcc.Get(fmt.Sprintf("pccserver/app-credentials/type/%s", serviceType), &profiles)
	return
}

// Get metadata app credentials
func (pcc *PccClient) GetMetadataAppCredentials(serviceType string) (profiles []authentication.AuthProfile, err error) {
	err = pcc.Get(fmt.Sprintf("pccserver/app-credentials/metadata/type/%s", serviceType), &profiles)
	return
}

// Get a single app credential
func (pcc *PccClient) GetAppCredential(id uint64) (profile *authentication.AuthProfile, err error) {
	err = pcc.Get(fmt.Sprintf("pccserver/app-credentials/%d", id), &profile)
	return
}

// Remove a credential profile (metadata or not)
func (pcc *PccClient) DeleteAppCredential(id uint64) (profile *authentication.AuthProfile, err error) {
	err = pcc.Delete(fmt.Sprintf("pccserver/app-credentials/%d", id), nil, nil)
	return
}

// Get a service template
func (pcc *PccClient) GetAppCredentialTemplate(serviceType string) (description authentication.AuthenticationServiceDescription, err error) {
	err = pcc.Get(fmt.Sprintf("pccserver/app-credentials/template/%s", serviceType), &description)
	return
}

// Returns all the supported services
func (pcc *PccClient) GetAppCredentialServices() (types []string, err error) {
	err = pcc.Get("pccserver/app-credentials/types", &types)
	return
}

// Create a profile (metadata or not)
func (pcc *PccClient) CreateAppCredentialProfile(profile *authentication.AuthProfile) (created authentication.AuthProfile, err error) {
	var b []byte
	m := make(map[string]string)
	if b, err = json.Marshal(profile); err == nil {
		m["data"] = string(b)
		err = pcc.PutFile("pccserver/app-credentials", "", m, &created)
	}
	return
}

// Update a profile (metadata or not)
func (pcc *PccClient) UpdateAppCredentialProfile(profile *authentication.AuthProfile) (created authentication.AuthProfile, err error) {
	var b []byte
	m := make(map[string]string)
	if b, err = json.Marshal(profile); err == nil {
		m["data"] = string(b)
		err = pcc.PutFiles("PUT", fmt.Sprintf("pccserver/app-credentials/%d", profile.Id), nil, m, &created)
	}
	return
}
