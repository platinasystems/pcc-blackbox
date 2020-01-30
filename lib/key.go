// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type Certificate struct {
	Id          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Alias       string `json:"alias"`
	Owner       uint64 `json:"owner"`
	Tenant      uint64 `json:"tenant"`
	Protect     bool   `json:"protect"`
}

type SecurityKey struct {
	Id          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Alias       string `json:"alias"`
	Type        string `json:"type"`
	Tenant      uint64 `json:"tenant"`
	Protect     bool   `json:"protect"`
	PrivatePath string `json:"privatePath"`
	PublicPath  string `json:"PublicPath"`
}

func (p PccClient) UploadKey(filePath string, label string, description string) (err error) {

	url := fmt.Sprintf("https://%s:9999/key-manager/keys/upload/public/%v",
		p.pccIp, label)

	file, err := os.Open(filePath)
	if err != nil {
		return
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("description", description)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return
	}
	err = writer.Close()
	if err != nil {
		return
	}
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Add("Authorization", p.bearer)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r, err := client.Do(req)
	defer r.Body.Close()
	return
}

func (p PccClient) DeleteKey(label string) (err error) {
	var (
		resp     HttpResp
		endpoint string
	)
	endpoint = fmt.Sprintf("key-manager/keys/%v", label)
	resp, _, err = p.pccSecurity("DELETE", endpoint, nil)
	if err != nil {
		return
	}
	if resp.Status != 200 {
		err = fmt.Errorf("%v", resp.Error)
		return
	}
	return
}

func (p PccClient) GetSecurityKeys() (secKeys []SecurityKey, err error) {
	var (
		resp     HttpResp
		endpoint string
	)
	endpoint = fmt.Sprintf("key-manager/keys/describe")
	resp, _, err = p.pccSecurity("GET", endpoint, nil)
	if err != nil {
		return
	}
	if resp.Status != 200 {
		err = fmt.Errorf("%v", resp.Error)
		return
	}
	err = json.Unmarshal(resp.Data, &secKeys)
	return
}
