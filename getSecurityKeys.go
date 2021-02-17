package main

import (
	"fmt"
	"testing"

	log "github.com/platinasystems/go-common/logs"
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
		log.AuctaLogger.Errorf("Error in retrieving Security Keys: %v\n", err)
		assert.FailNow()
		return
	}

	for i := 0; i < len(secKeys); i++ {
		SecurityKeys[secKeys[i].Alias] = &secKeys[i]
		log.AuctaLogger.Infof("Mapping SecurityKey[%v]:%d - %v\n",
			secKeys[i].Alias, secKeys[i].Id, secKeys[i].Description)
	}
}

func getFirstKey() (sKey pcc.SecurityKey, err error) {
	var secKeys []pcc.SecurityKey
	if secKeys, err = Pcc.GetSecurityKeys(); err == nil {
		if len(secKeys) == 0 {
			err = fmt.Errorf("key not found")
		} else {
			sKey = secKeys[0]
		}

	}

	return
}
