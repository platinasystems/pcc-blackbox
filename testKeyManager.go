package main

import (
	"fmt"
	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/pcc-blackbox/models"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
	"time"
)

func testKMKeys(t *testing.T) {

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	log.AuctaLogger.Info("KEYS:")
	var (
		key pcc.SecurityKey
		err error
	)

	alias := "key-bb-test"
	description := "key-bb-test"

	Pcc.DeleteKey(alias) // delete if exist
	fileName := fmt.Sprintf("%s.pem", alias)
	log.AuctaLogger.Info(fmt.Sprintf("generating the key %s %s", alias, fileName))
	cmd := exec.Command("/usr/bin/openssl", "genrsa", "-out", fileName, "2048")
	if err = cmd.Run(); err != nil {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

	log.AuctaLogger.Infof("uploading the key", alias)
	if _, err = Pcc.UploadKey(fileName, alias, pcc.PRIVATE_KEY, description); err == nil { // TODO check if the key already exist
		log.AuctaLogger.Infof("Added the key", alias)
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

	defer func() {
		log.AuctaLogger.Infof("deleting the key", alias)
		err = Pcc.DeleteKey(alias) // delete at the end
		if err != nil {
			log.AuctaLogger.Infof("Delete key [%v] failed: %v", alias, err)
			msg := fmt.Sprintf("%v", err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		}
	}()

	log.AuctaLogger.Infof("comparing the content for the key", alias)
	if content, err := Pcc.DownloadKey(alias, pcc.PRIVATE_KEY); err == nil { // compare the content
		readFileAndCompare(t, content, fileName)
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

	log.AuctaLogger.Info(fmt.Sprintf("looking for the key %s", alias))
	if items, err := Pcc.GetSecurityKeys(); err == nil {
		for _, c := range items {
			if c.Alias == alias {
				goto cont
			}
		}
		msg := fmt.Sprintf("not able to found the key %s", alias)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

cont:
	log.AuctaLogger.Infof("getting the key", alias)
	if key, err = Pcc.GetSecurityKey(alias); err == nil {
		if alias != key.Alias || description != key.Description || key.Protect {
			msg := fmt.Sprintf("the describe returned some different values %v", key)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		}
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

	log.AuctaLogger.Infof("updating the key", key.Alias)
	previous := key.Description
	key.Description = key.Description + "new"
	if err := Pcc.UpdateSecurityKey(key); err == nil {
		if previous == key.Description {
			msg := fmt.Sprintf("the description does not change for the key %s", key.Alias)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		}
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}
}

func testKMCertificates(t *testing.T) {

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	log.AuctaLogger.Info("CERTIFICATES:")
	var (
		err  error
		cert *pcc.Certificate
	)

	alias := "certificate-bb-test"
	description := "blackbox certificate"
	log.AuctaLogger.Info(fmt.Sprintf("looking for the certificate %s and deleting if exists", alias))
	if items, err := Pcc.GetCertificates(); err == nil {
		for _, c := range items {
			if c.Alias == alias {
				err = Pcc.DeleteCertificate(c.Id)
				if err != nil {
					log.AuctaLogger.Infof("Delete cert failed: %v",
						err)
				}
				break
			}
		}
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

	fileName := fmt.Sprintf("%s.crt", alias)
	keyName := fmt.Sprintf("%s.pem", alias)
	log.AuctaLogger.Info(fmt.Sprintf("generating the certificate %s %s", alias, fileName))
	cmd := exec.Command("/usr/bin/openssl", "req", "-nodes", "-new", "-x509", "-keyout", keyName, "-out", fileName, "--subj", "/C=US/ST=SanFrancisco/L=SanFrancisco/O=Global Security/OU=IT Department/CN=platinasystems.net")
	if err = cmd.Run(); err != nil {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

	log.AuctaLogger.Infof("uploading the certificate ", fileName)
	if _, err = Pcc.UploadCert(fileName, alias, description, 0); err != nil {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

	log.AuctaLogger.Infof("uploaded the certificate", alias)
	if certs, err := Pcc.GetCertificates(); err == nil { // Look for the certificate
		for i := range certs {
			if certs[i].Alias == alias {
				cert = &certs[i]
				goto CONT
			}
		}
		msg := "unable to find the certificate"
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

CONT:
	defer func() {
		if cert != nil {
			log.AuctaLogger.Infof("deleting the certificate", cert.Id, cert.Alias)
			err = Pcc.DeleteCertificate(cert.Id) // delete at the end
			if err != nil {
				msg := fmt.Sprintf("Delete cert failed: %v", err)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				t.FailNow()
			}
		}
	}()

	if c, err := Pcc.GetCertificate(cert.Id); err == nil {
		if c.Alias != alias || c.Protect || c.Description != description {
			msg := fmt.Sprintf("the describe returned some different values %v , %v", c, cert)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		}
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}

	log.AuctaLogger.Infof("comparing the content for the certificate", *cert)
	if content, err := Pcc.DownloadCertificate(cert.Id); err == nil { // compare the content
		readFileAndCompare(t, content, fileName)
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}
}

// read from file and compare the content
func readFileAndCompare(t *testing.T, content string, fileName string) {

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	if file, err := os.Open(fileName); err == nil {
		defer file.Close()
		if b, err := ioutil.ReadAll(file); err == nil {
			if string(b) != content {
				msg := fmt.Sprintf("the downloaded file is different from %s", fileName)
				res.SetTestFailure(msg)
				log.AuctaLogger.Error(msg)
				t.FailNow()
			}
		} else {
			msg := fmt.Sprintf("%v", err)
			res.SetTestFailure(msg)
			log.AuctaLogger.Error(msg)
			t.FailNow()
		}
	} else {
		msg := fmt.Sprintf("%v", err)
		res.SetTestFailure(msg)
		log.AuctaLogger.Error(msg)
		t.FailNow()
	}
}
