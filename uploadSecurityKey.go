package main

import (
	"fmt"
	"os"
	"testing"

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
	assert := test.Assert{t}
	f, err := os.OpenFile("maas_pubkey", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		log.AuctaLogger.Errorf("Unable to create file:\n%v\n", err)
		assert.FailNow()
		return
	}
	_, err = f.Write([]byte(PUB_KEY))
	if err != nil {
		log.AuctaLogger.Errorf("Unable to write on disk:\n%v\n", err)
		assert.FailNow()
		return
	}
	f.Close()

	for i := 0; ; i++ {
		label := fmt.Sprintf("test_%d", i)
		description := "Don't be evil"
		exist, err := Pcc.CheckKeyLabelExists(label)
		if err != nil {
			log.AuctaLogger.Errorf("%v\n", err)
			assert.FailNow()
		}
		if !exist {
			_, err = Pcc.UploadKey("./maas_pubkey", label,
				pcc.PUBLIC_KEY, description)
			if err != nil {
				log.AuctaLogger.Errorf("%v\n", err)
				assert.FailNow()
				return
			}
			break
		}
	}
}

func delAllKeys(t *testing.T) {
	test.SkipIfDryRun(t)
	assert := test.Assert{t}
	var (
		secKeys []pcc.SecurityKey
		err     error
	)

	secKeys, err = Pcc.GetSecurityKeys()
	if err != nil {
		log.AuctaLogger.Errorf("Failed to GetSecurityKeys: %v\n", err)
		assert.FailNow()
		return
	}

	for i := 0; i < len(secKeys); i++ {
		err = Pcc.DeleteKey(secKeys[i].Alias)
		if err != nil {
			log.AuctaLogger.Errorf("Failed to delete key %v: %v\n",
				secKeys[i].Alias, err)
			assert.FailNow()
			return
		}
	}
}
