// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import "fmt"

type Credential struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

type PccClient struct {
	pccIp  string
	bearer string
}

// GET
func (pcc *PccClient) Get(endPoint string, out interface{}) (err error) {
	err = (&RestClient{address: pcc.pccIp, token: pcc.bearer}).Get(endPoint, out)
	return
}

// DELETE
func (pcc *PccClient) Delete(endPoint string, out interface{}) (err error) {
	err = (&RestClient{address: pcc.pccIp, token: pcc.bearer}).Delete(endPoint, out)
	return
}

// POST
func (pcc *PccClient) Post(endPoint string, data interface{}, out interface{}) (err error) {
	err = (&RestClient{address: pcc.pccIp, token: pcc.bearer}).Post(endPoint, data, out)
	return
}

// PUT
func (pcc *PccClient) Put(endPoint string, data interface{}, out interface{}) (err error) {
	err = (&RestClient{address: pcc.pccIp, token: pcc.bearer}).Put(endPoint, data, out)
	return
}

// PUT
func (pcc *PccClient) PutFile(endPoint string, filePath string, fields map[string]string, out interface{}) (err error) { // FIXME all services should share the same structure
	err = (&RestClient{address: pcc.pccIp, token: pcc.bearer}).PutFile(endPoint, filePath, fields, out)
	return
}

func (pcc *PccClient) GetFile(endPoint string) (content string, err error) {
	content, err = (&RestClient{address: pcc.pccIp, token: pcc.bearer}).GetFile(endPoint)
	return
}

func Authenticate(PccIp string, cred Credential) (pcc *PccClient, err error) {
	var out struct{ Token string }
	rc := RestClient{address: PccIp}
	client := PccClient{pccIp: PccIp}
	if err = rc.Post("security/auth", cred, &out); err == nil {
		client.bearer = fmt.Sprintf("Bearer %s", out.Token)
	}

	pcc = &client
	return
}
