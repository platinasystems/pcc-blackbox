package main

import (
	"fmt"
	log "github.com/platinasystems/go-common/logs"
	"github.com/platinasystems/pcc-blackbox/models"
	"os"
	"testing"
	"time"

	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/test"
)

func CreateFileAndUpload(fileName string, key string, fileType string, keyId uint64) (err error) {
	var f *os.File
	f, err = os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		err = fmt.Errorf("Unable to create file:\n%v\n", err)
		return
	}
	defer f.Close()
	_, err = f.Write([]byte(key))
	if err != nil {
		err = fmt.Errorf("Unable to write on disk:\n%v\n", err)
		return
	}

	filePath := fmt.Sprintf("./%s", fileName)

	// check if exist and delete if so
	switch fileType {
	case pcc.PRIVATE_KEY, pcc.PUBLIC_KEY:
		var (
			exist bool
		)
		exist, _, err = Pcc.FindSecurityKey(fileName)
		if exist {
			Pcc.DeleteKey(fileName)
		}

		_, err = Pcc.UploadKey(filePath, fileName, fileType, "")
		if err != nil {
			return
		}
	case pcc.CERT:
		var (
			cert  pcc.Certificate
			exist bool
		)
		exist, cert, err = Pcc.FindCertificate(fileName)
		if exist {
			Pcc.DeleteCertificate(cert.Id)
		}

		_, err = Pcc.UploadCert(filePath, fileName, "", keyId)
		if err != nil {
			return
		}
	}

	return
}

func delAllCerts(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now(), "delAllCerts")
	assert := test.Assert{t}

	var (
		certificates []pcc.Certificate
		id           uint64
		err          error
	)

	certificates, err = Pcc.GetCertificates()
	if err != nil {
		msg := fmt.Sprintf("Failed to get certificates: %v\n", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		assert.FailNow()
		return
	}

	for _, c := range certificates {
		id = c.Id
		log.AuctaLogger.Infof("Deleting certificate %v\n", c.Alias)
		err = Pcc.DeleteCertificate(id)
		if err != nil {
			msg := fmt.Sprintf("Failed to delete Certificate %v: %v\n",
				id, err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			assert.FailNow()
			return
		}
	}
}
