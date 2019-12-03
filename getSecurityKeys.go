package main

import (
	"encoding/json"
	"fmt"
	"github.com/platinasystems/test"
	"testing"
)

type securityKey struct {
	Id          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Alias       string `json:"alias"`
	Type        string `json:"type"`
	Tenant      uint64 `json:"tenant"`
	Protect     bool   `json:"protect"`
	PrivatePath string `json:"privatePath"`
	PublicPath  string `json:"PublicPath"`
}

func getSecurityKeys(t *testing.T) {
	t.Run("getSecKeys", getSecKeys)
}

func getSecKeys(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var secKeys []securityKey

	resp, err := getSecurityKeyLists()
	if err != nil {
		assert.Fatalf("Error in retrieving Security Keys:\n%v\n%v\n", string(resp.Message), err)
		return
	}
	if err := json.Unmarshal(resp.Data, &secKeys); err != nil {
		assert.Fatalf("%v\n%v\n", string(resp.Data), err)
		return
	}
	for i := 0; i < len(secKeys); i++ {
		SecurityKeys[secKeys[i].Alias] = &secKeys[i]
		fmt.Printf("Mapping SecurityKey[%v]:%d\n", secKeys[i].Alias, secKeys[i].Id)
	}
}

func getFirstKey() (sKey securityKey, err error) {
	var secKeys []securityKey
	resp, err := getSecurityKeyLists()
	if err != nil {
		return sKey, err
	}
	json.Unmarshal(resp.Data, &secKeys)
	if err != nil {
		return sKey, err
	}
	if len(secKeys) == 0 {
		return sKey, err
	}
	return secKeys[0], err
}

func getSecurityKeyLists() (resp HttpResp, err error) {
	var body []byte
	endpoint := fmt.Sprintf("key-manager/keys/describe")
	resp, body, err = pccSecurity("GET", endpoint, nil)
	if err != nil {
		fmt.Printf("Get keys list failed\n%v\n", string(body))
		return resp, err
	}
	return resp, err
}
