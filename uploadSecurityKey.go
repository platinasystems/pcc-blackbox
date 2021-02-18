package main

import (
	"fmt"
	"github.com/platinasystems/pcc-blackbox/models"
	"os"
	"testing"
	"time"

	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

func updateSecurityKey(t *testing.T) {
	t.Run("updateSecurityKey", updateSecurityKey_MaaS)
}

const PUB_KEY = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC0ebbgMSADWEvLIDWnusBsPLdzeHHeeHAxahRuwQPfhzrGctIMM40wMG7TQ4hz/hL7FAxsIuJseG8Q3LFOHfW7W0tLMLwilQgd4lqZm7RBjFJ+zoWsw1wJIYDsqlxZiFxzffntRwpX7giz9CJZ9h9qDgimeWbClO4Gr2h99UcWbYtnzZYy/eHOpYX4yZrluQvN9guGjrClcFa9Ye4Ayq93wgiSHbFuOC0gqR0JqO8/tJ4dctQ1OPLddLRKtJ0YuKL6bgDtrqGlTsnXeOR0lzjFXhNVAfEtcMFLFDDpLaoquqRiWYtgLI5RJHwOLI3YFE02qNWxBs9WQe2AaYw4fBc3 gmorana@Giovannis-MBP.homenet.telecomitalia.it"

type secKeyUploader struct {
	Key os.File
}

func updateSecurityKey_MaaS(t *testing.T) {
	test.SkipIfDryRun(t)
	res := models.InitTestResult(runID)
	defer res.SaveTestResult()
	defer res.SetElapsedTime(time.Now(), "updateSecurityKey_MaaS")
	assert := test.Assert{t}
	f, err := os.OpenFile("maas_pubkey", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		msg := fmt.Sprintf("Unable to create file:\n%v\n", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}
	_, err = f.Write([]byte(PUB_KEY))
	if err != nil {
		msg := fmt.Sprintf("Unable to write on disk:\n%v\n", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}
	f.Close()

	for i := 0; ; i++ {
		label := fmt.Sprintf("test_%d", i)
		description := "Don't be evil"
		exist, err := Pcc.CheckKeyLabelExists(label)
		if err != nil {
			msg := fmt.Sprintf("%v\n", err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
		}
		if !exist {
			_, err = Pcc.UploadKey("./maas_pubkey", label,
				pcc.PUBLIC_KEY, description)
			if err != nil {
				msg := fmt.Sprintf("%v\n", err)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				assert.FailNow()
				return
			}
			break
		}
	}
	res.SetTestPass()
}

func delAllKeys(t *testing.T) {
	test.SkipIfDryRun(t)
	res := models.InitTestResult(runID)
	defer res.SaveTestResult()
	defer res.SetElapsedTime(time.Now(), "delAllKeys")
	assert := test.Assert{t}
	var (
		secKeys []pcc.SecurityKey
		err     error
	)

	secKeys, err = Pcc.GetSecurityKeys()
	if err != nil {
		msg := fmt.Sprintf("Failed to GetSecurityKeys: %v\n", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}

	for i := 0; i < len(secKeys); i++ {
		err = Pcc.DeleteKey(secKeys[i].Alias)
		if err != nil {
			msg := fmt.Sprintf("Failed to delete key %v: %v\n",
				secKeys[i].Alias, err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
			return
		}
	}
	res.SetTestPass()
}
