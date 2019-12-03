package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"

	//"encoding/json"
	"fmt"
	"github.com/platinasystems/test"
	"os"
	"testing"
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
		assert.Fatalf("Unable to create file:\n%v\n", err)
		return
	}
	_, err = f.Write([]byte(PUB_KEY))
	if err != nil {
		assert.Fatalf("Unable to write on disk:\n%v\n", err)
		return
	}
	f.Close()

	for i := 0; ; i++ {
		label := fmt.Sprintf("test_%d", i)
		exist, err := checkIfLabelExist(label)
		if err != nil {
			assert.Fatalf("%v\n", err)
		}
		if !exist {
			err = updateFile("./maas_pubkey", label)
			if err != nil {
				assert.Fatalf("%v\n", err)
				return
			}
			break
		}
	}
}

func checkIfLabelExist(label string) (exist bool, err error) {
	var secKeys []securityKey
	resp, err := getSecurityKeyLists()
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(resp.Data, &secKeys); err != nil {
		fmt.Printf("%v\n%v\n", string(resp.Data), err)
		return false, err
	}
	for i := 0; i < len(secKeys); i++ {
		//update
		SecurityKeys[secKeys[i].Alias] = &secKeys[i]
		if secKeys[i].Alias == label {
			exist = true
		}
	}
	return exist, err
}

func updateFile(filePath string, label string) (err error) {
	file, err := os.Open(filePath)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base("./maas_pubkey"))
	if err != nil {
		return nil
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil
	}
	err = writer.Close()
	if err != nil {
		return nil
	}
	client := &http.Client{}
	url := fmt.Sprintf("https://%s:9999/key-manager/keys/upload/public/%v", Env.PccIp, label)
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Add("Authorization", Bearer)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r, err := client.Do(req)
	defer r.Body.Close()
	return err
}
