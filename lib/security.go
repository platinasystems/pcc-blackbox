// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Credential struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

type PccClient struct {
	pccIp  string
	bearer string
}

func Authenticate(PccIp string, cred Credential) (pcc PccClient, err error) {
	var (
		data []byte
		resp *http.Response
	)

	data, err = json.Marshal(cred)

	url := fmt.Sprintf("https://%s:9999/security/auth", PccIp)
	http.DefaultTransport.(*http.Transport).TLSClientConfig =
		&tls.Config{InsecureSkipVerify: true}

	resp, err = http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var out struct{ Token string }
	err = json.Unmarshal(body, &out)
	if err != nil {
		return
	}
	pcc.bearer = "Bearer " + out.Token
	pcc.pccIp = PccIp
	return
}
