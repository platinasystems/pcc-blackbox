package main

import (
	"fmt"

	"testing"

	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

func getSecurityKeys(t *testing.T) {
	t.Run("getSecKeys", getSecKeys)
}

func getSecKeys(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}

	var (
		secKeys []pcc.SecurityKey
		err     error
	)

	secKeys, err = Pcc.GetSecurityKeys()
	if err != nil {
		assert.Fatalf("Error in retrieving Security Keys: %v\n", err)
		return
	}

	for i := 0; i < len(secKeys); i++ {
		SecurityKeys[secKeys[i].Alias] = &secKeys[i]
		fmt.Printf("Mapping SecurityKey[%v]:%d - %v\n",
			secKeys[i].Alias, secKeys[i].Id, secKeys[i].Description)
	}
}

func getFirstKey() (sKey pcc.SecurityKey, err error) {

	var secKeys []pcc.SecurityKey

	secKeys, err = Pcc.GetSecurityKeys()
	if err != nil {
		return
	}

	return secKeys[0], err
}
