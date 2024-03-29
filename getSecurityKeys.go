package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/platinasystems/pcc-blackbox/models"

	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

func getSecurityKeys(t *testing.T) {
	t.Run("getSecKeys", getSecKeys)
}

func getSecKeys(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	assert := test.Assert{t}

	var (
		secKeys []pcc.SecurityKey
		err     error
	)

	secKeys, err = Pcc.GetSecurityKeys()
	if err != nil {
		msg := fmt.Sprintf("Error in retrieving Security Keys: %v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
	}

	for i := 0; i < len(secKeys); i++ {
		SecurityKeys[secKeys[i].Alias] = &secKeys[i]
		log.AuctaLogger.Infof("Mapping SecurityKey[%v]:%d - %v",
			secKeys[i].Alias, secKeys[i].Id, secKeys[i].Description)
	}
}

func getFirstKey() (sKey pcc.SecurityKey, err error) {
	var secKeys []pcc.SecurityKey
	if secKeys, err = Pcc.GetSecurityKeys(); err == nil {
		if len(secKeys) == 0 {
			err = fmt.Errorf("key not found")
		} else {
			for _, key := range secKeys {
				if key.HasPublic {
					sKey = key
					break
				}
			}
			if sKey.Id == 0 {
				err = fmt.Errorf("key not found")
			}
		}
	}

	return
}
