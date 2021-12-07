package main

import (
	"errors"
	"fmt"
	log "github.com/platinasystems/go-common/logs"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/pcc-blackbox/models"
	"github.com/platinasystems/test"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
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
	err = cmd.Run()
	checkError(t, res, err)

	log.AuctaLogger.Infof("uploading the key", alias)
	_, err = Pcc.UploadKey(fileName, alias, pcc.PRIVATE_KEY, description)
	checkError(t, res, err)
	log.AuctaLogger.Infof("Added the key", alias)

	defer func() {
		log.AuctaLogger.Infof("deleting the key", alias)
		err = Pcc.DeleteKey(alias) // delete at the end
		checkError(t, res, err)
	}()

	log.AuctaLogger.Infof("comparing the content for the key", alias)
	content, err := Pcc.DownloadKey(alias, pcc.PRIVATE_KEY)
	checkError(t, res, err)
	readFileAndCompare(t, content, fileName)

	log.AuctaLogger.Info(fmt.Sprintf("looking for the key %s", alias))
	items, err := Pcc.GetSecurityKeys()
	checkError(t, res, err)
	for _, c := range items {
		if c.Alias == alias {
			goto cont
		}
	}

	err = errors.New(fmt.Sprintf("not able to found the key %s", alias))
	checkError(t, res, err)

cont:
	log.AuctaLogger.Infof("getting the key", alias)
	key, err = Pcc.GetSecurityKey(alias)
	checkError(t, res, err)

	if alias != key.Alias || description != key.Description || key.Protect {
		err = errors.New(fmt.Sprintf("the describe returned some different values %v", key))
		checkError(t, res, err)
	}

	log.AuctaLogger.Infof("updating the key", key.Alias)
	previous := key.Description
	key.Description = key.Description + "new"
	err = Pcc.UpdateSecurityKey(key)
	checkError(t, res, err)

	if previous == key.Description {
		err = errors.New(fmt.Sprintf("the description does not change for the key %s", key.Alias))
		checkError(t, res, err)
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
	items, err := Pcc.GetCertificates()
	checkError(t, res, err)

	for _, c := range items {
		if c.Alias == alias {
			err = Pcc.DeleteCertificate(c.Id)
			checkError(t, res, err)
			log.AuctaLogger.Infof("Delete cert failed: %v", err)
			break
		}
	}

	fileName := fmt.Sprintf("%s.crt", alias)
	keyName := fmt.Sprintf("%s.pem", alias)
	log.AuctaLogger.Info(fmt.Sprintf("generating the certificate %s %s", alias, fileName))
	cmd := exec.Command("/usr/bin/openssl", "req", "-nodes", "-new", "-x509", "-keyout", keyName, "-out", fileName, "--subj", "/C=US/ST=SanFrancisco/L=SanFrancisco/O=Global Security/OU=IT Department/CN=platinasystems.net")
	err = cmd.Run()
	checkError(t, res, err)

	log.AuctaLogger.Infof("uploading the certificate ", fileName)
	_, err = Pcc.UploadCert(fileName, alias, description, 0)
	checkError(t, res, err)

	log.AuctaLogger.Infof("uploaded the certificate", alias)
	certs, err := Pcc.GetCertificates()
	checkError(t, res, err)
	// Look for the certificate
	for i := range certs {
		if certs[i].Alias == alias {
			cert = &certs[i]
			goto CONT
		}
	}
	err = errors.New("unable to find the certificate")
	checkError(t, res, err)

CONT:
	defer func() {
		if cert != nil {
			log.AuctaLogger.Infof("deleting the certificate", cert.Id, cert.Alias)
			err = Pcc.DeleteCertificate(cert.Id) // delete at the end
			checkError(t, res, err)
		}
	}()

	c, err := Pcc.GetCertificate(cert.Id)
	checkError(t, res, err)
	if c.Alias != alias || c.Protect || c.Description != description {
		err = errors.New(fmt.Sprintf("the describe returned some different values %v , %v", c, cert))
		checkError(t, res, err)
	}

	log.AuctaLogger.Infof("comparing the content for the certificate", *cert)
	content, err := Pcc.DownloadCertificate(cert.Id)
	checkError(t, res, err)
	// compare the content
	readFileAndCompare(t, content, fileName)
}

func addPrivatePublicCert(t *testing.T) {
	test.SkipIfDryRun(t)

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	alias := "test-cert-bb"
	description := "private/public blackbox certificate"

	fileName := "cert-bb.pem"
	keyName := "key-bb.pem"

	exists, _, err := Pcc.FindCertificate(alias)
	checkError(t, res, err)
	if exists {
		log.AuctaLogger.Warnf("A certificate with name %s already exists and will be used", alias)
		return
	}

	_, err = Pcc.UploadCertPrivatePublic(fileName, keyName, alias, description)
	checkError(t, res, err)
}

// read from file and compare the content
func readFileAndCompare(t *testing.T, content string, fileName string) {

	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	file, err := os.Open(fileName)
	checkError(t, res, err)
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	checkError(t, res, err)

	trimmedB, trimmedContent := strings.TrimSuffix(string(b), "\n"), strings.TrimSuffix(content, "\n")
	if trimmedB != trimmedContent {
		err = errors.New(fmt.Sprintf("the downloaded file is different from %s", fileName))
		checkError(t, res, err)
	}
}
