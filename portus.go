package main

import (
	"encoding/json"
	"fmt"
	"github.com/platinasystems/test"
	"testing"
	"github.com/platinasystems/tiles/pccserver/models"
	"time"
)

const (
	PORTUS_ENDPOINT      = "pccserver/portus"
	KEYMANAGER_ENDPOINT  = "key-manager"
	PORTUS_KEY_FILENAME  = "test_portus_key"
	PORTUS_CERT_FILENAME = "test_portus_crt"
)

var PortusSelectedNodeId uint64

func AddPortus(t *testing.T) {
	t.Run("addPortus", installPortus)
}

func CheckPortusInstallation(t *testing.T) {
	t.Run("checkPortus", checkPortus)
}

func UploadSecurityPortusKey(t *testing.T) {
	t.Run("uploadPortusKey", uploadSecurityKey_Portus)
}

func UploadSecurityPortusCert(t *testing.T) {
	t.Run("uploadSecurityPortusCert", uploadCertificate_Portus)
}

func uploadSecurityKey_Portus(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	err := CreateFileAndUpload(PORTUS_KEY_FILENAME, PORTUS_KEY, PRIVATE_KEY)
	if err != nil {
		assert.Fatalf(err.Error())
	}
}

func uploadCertificate_Portus(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	err := CreateFileAndUpload(PORTUS_CERT_FILENAME, PORTUS_CERT, CERT)
	if err != nil {
		assert.Fatalf(err.Error())
	}
}

func installPortus(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var (
		portusConfiguration models.PortusConfiguration
		body                []byte
		resp                HttpResp
	)
	for id, node := range Nodes {
		if !IsInvader(node) && IsOnline(node) {
			portusConfiguration = Env.PortusConfiguration
			portusConfiguration.NodeID = id
			portusConfiguration.Name = fmt.Sprintf("portus_%v", id)

			if Env.AuthenticationProfile.Name == "" {
				fmt.Printf("Authenticatiom Profile is not defined in the configuration file, Portus will be installed without it")
			} else {
				authProfile, err := GetAuthProfileByName(CurrentAuthProfileName)
				if err == nil {
					portusConfiguration.AuthenticationProfile = authProfile
				} else {
					fmt.Printf("Missing authentication profile %s\n, Portus will be installed without it", CurrentAuthProfileName)
				}
			}

			certificate, err := GetCertificate(PORTUS_CERT_FILENAME)
			if err != nil {
				fmt.Printf("Get certificate %s failed\n%v\n", PORTUS_CERT_FILENAME, err)
			} else {
				portusConfiguration.RegistryCertId = &certificate.Id
			}

			privateKey, err := GetSecurityKey(PORTUS_KEY_FILENAME, PRIVATE_KEY)
			if err != nil {
				fmt.Printf("Get private key %s failed\n%v\n", PORTUS_KEY_FILENAME, err)
			} else {
				portusConfiguration.RegistryKeyId = &privateKey.Id
			}

			data, err := json.Marshal(portusConfiguration)
			if err != nil {
				assert.Fatalf("invalid struct for install portus request")
			}
			fmt.Println(fmt.Sprintf("Installing Portus on Node with id %v", node.Id))
			PortusSelectedNodeId = node.Id
			if resp, body, err = pccGateway("POST", PORTUS_ENDPOINT, data); err != nil {
				assert.Fatalf("%v\n%v\n", string(body), err)
				return
			}
			if resp.Status != 200 {
				assert.Fatalf("%v\n", string(body))
				fmt.Printf("install Portus %v failed\n%v\n", node.Host, string(body))
				return
			}
			break
		}

	}
}

func checkPortus(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	from := time.Now()

	for id, node := range Nodes {
		if !IsInvader(node) && IsOnline(node) && node.Id == PortusSelectedNodeId{
			check, err := checkGenericInstallation(id, PORTUS_TIMEOUT, PORTUS_NOTIFICATION, from)
			if err != nil {
				assert.Fatalf("Portus installation has failed\n%v\n", err)
			}
			if check {
				fmt.Printf("Portus correctly installed on nodeId:%v\n", id)
			}
		}
	}
}
