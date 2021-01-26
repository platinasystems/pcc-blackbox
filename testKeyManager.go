package main

import (
	"fmt"
	pcc "github.com/platinasystems/pcc-blackbox/lib"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

func testKMKeys(t *testing.T) {
	fmt.Printf("\n\nKEYS:\n")
	var (
		key pcc.SecurityKey
		err error
	)

	alias := "key-bb-test"
	description := "key-bb-test"

	Pcc.DeleteKey(alias) // delete if exist
	fileName := fmt.Sprintf("%s.pem", alias)
	fmt.Println(fmt.Sprintf("generating the key %s %s", alias, fileName))
	cmd := exec.Command("/usr/bin/openssl", "genrsa", "-out", fileName, "2048")
	if err = cmd.Run(); err != nil {
		t.Fatal(err)
	}

	fmt.Println("uploading the key", alias)
	if _, err = Pcc.UploadKey(fileName, alias, pcc.PRIVATE_KEY, description); err == nil { // TODO check if the key already exist
		fmt.Println("Added the key", alias)
	} else {
		t.Fatal(err)
	}

	defer func() {
		fmt.Println("deleting the key", alias)
		Pcc.DeleteKey(alias) // delete at the end
	}()

	fmt.Println("comparing the content for the key", alias)
	if content, err := Pcc.DownloadKey(alias, pcc.PRIVATE_KEY); err == nil { // compare the content
		readFileAndCompare(t, content, fileName)
	} else {
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf("looking for the key %s", alias))
	if items, err := Pcc.GetSecurityKeys(); err == nil {
		for _, c := range items {
			if c.Alias == alias {
				goto cont
			}
		}
		t.Fatal(fmt.Sprintf("not able to found the key %s", alias))
	} else {
		t.Fatal(err)
	}

cont:
	fmt.Println("getting the key", alias)
	if key, err = Pcc.GetSecurityKey(alias); err == nil {
		if alias != key.Alias || description != key.Description || key.Protect {
			t.Fatal("the describe returned some different values", key)
		}
	} else {
		t.Fatal(err)
	}

	fmt.Println("updating the key", key.Alias)
	previous := key.Description
	key.Description = key.Description + "new"
	if err := Pcc.UpdateSecurityKey(key); err == nil {
		if previous == key.Description {
			t.Fatalf("the description does not change for the key %s", key.Alias)
		}
	} else {
		t.Fatal(err)
	}
}

func testKMCertificates(t *testing.T) {
	fmt.Printf("\n\nCERTIFICATES:\n")
	var (
		err  error
		cert *pcc.Certificate
	)

	alias := "certificate-bb-test"
	description := "blackbox certificate"
	fmt.Println(fmt.Sprintf("looking for the certificate %s and deleting if exists", alias))
	if items, err := Pcc.GetCertificates(); err == nil {
		for _, c := range items {
			if c.Alias == alias {
				Pcc.DeleteCertificate(c.Id)
				break
			}
		}
	} else {
		t.Fatal(err)
	}

	fileName := fmt.Sprintf("%s.crt", alias)
	keyName := fmt.Sprintf("%s.pem", alias)
	fmt.Println(fmt.Sprintf("generating the certificate %s %s", alias, fileName))
	cmd := exec.Command("/usr/bin/openssl", "req", "-nodes", "-new", "-x509", "-keyout", keyName, "-out", fileName, "--subj", "/C=US/ST=SanFrancisco/L=SanFrancisco/O=Global Security/OU=IT Department/CN=platinasystems.net")
	if err = cmd.Run(); err != nil {
		t.Fatal(err)
	}

	fmt.Println("uploading the certificate", fileName)
	if _, err = Pcc.UploadCert(fileName, alias, description, 0); err != nil {
		t.Fatal(err)
	}

	fmt.Println("uploaded the certificate", alias)
	if certs, err := Pcc.GetCertificates(); err == nil { // Look for the certificate
		for i := range certs {
			if certs[i].Alias == alias {
				cert = &certs[i]
				goto CONT
			}
		}
		t.Fatal("unable to find the certificate")
	} else {
		t.Fatal(err)
	}

CONT:
	defer func() {
		if cert != nil {
			fmt.Println("deleting the certificate", cert.Id, cert.Alias)
			Pcc.DeleteCertificate(cert.Id) // delete at the end
		}
	}()

	if c, err := Pcc.GetCertificate(cert.Id); err == nil {
		if c.Alias != alias || c.Protect || c.Description != description {
			t.Fatal("the describe returned some different values", c, cert)
		}
	} else {
		t.Fatal(err)
	}

	fmt.Println("comparing the content for the certificate", *cert)
	if content, err := Pcc.DownloadCertificate(cert.Id); err == nil { // compare the content
		readFileAndCompare(t, content, fileName)
	} else {
		t.Fatal(err)
	}
}

// read from file and compare the content
func readFileAndCompare(t *testing.T, content string, fileName string) {
	if file, err := os.Open(fileName); err == nil {
		defer file.Close()
		if b, err := ioutil.ReadAll(file); err == nil {
			if string(b) != content {
				t.Fatal("the downloaded file is different from", fileName)
			}
		} else {
			t.Fatal(err)
		}
	} else {
		t.Fatal(err)
	}
}
