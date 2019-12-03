package main

import (
	"encoding/json"
	"fmt"
	"github.com/platinasystems/test"
	"io/ioutil"
	"testing"
	"tiles/pccserver/models"
)

const (
	PORTUS_ENDPOINT     = "pccserver/portus"
	PORTUS_JSON         = "authProfile.json"
	PORTUS_TIMEOUT      = 300
	PORTUS_NOTIFICATION = "[Portus] has been installed correctly"
)

func addPortus(t *testing.T) {
	t.Run("addPortus", installPortus)
}

func checkPortusInstallation(t *testing.T) {
	t.Run("checkPortus", checkPortus)
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

		if !node.Invader {
			if err := buildPortus(&portusConfiguration); err != nil {
				assert.Fatalf("Portus Configuration creation failed\n%v\n", err)
			}
			portusConfiguration.NodeID = id
			portusConfiguration.Name = fmt.Sprintf("portus_%v", id)

			data, err := json.Marshal(portusConfiguration)
			if err != nil {
				assert.Fatalf("invalid struct for install portus request")
			}

			if resp, body, err = pccGateway("POST", PORTUS_ENDPOINT, data); err != nil {
				assert.Fatalf("%v\n%v\n", string(body), err)
				return
			}
			if resp.Status != 200 {
				assert.Fatalf("%v\n", string(body))
				fmt.Printf("install Portus %v failed\n%v\n", node.Host, string(body))
				return
			}
		}

	}
}

func checkPortus(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	for id, node := range Nodes {
		if !node.Invader {
			check, err := checkGenericInstallation(id, PORTUS_TIMEOUT, PORTUS_NOTIFICATION)
			if err != nil {
				assert.Fatalf("Portus installation has failed\n%v\n", err)
			}
			if check {
				fmt.Printf("Portus correctly installed on nodeId:%v\n", id)
			}
		}
	}

}

func buildPortus(portusConfiguration *models.PortusConfiguration) (err error) {
	confFile, err := ioutil.ReadFile(PORTUS_JSON)
	json.Unmarshal(confFile, portusConfiguration)
	return err
}
