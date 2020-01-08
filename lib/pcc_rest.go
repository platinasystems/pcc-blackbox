// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type respGeneric struct {
	Path    string `json:"path"`
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error"`
	Data    interface{}
}

type HttpResp struct {
	Path    string
	Status  int
	Message string
	Error   string
	Data    []byte
}

func (p PccClient) pccGateway(op string, endPoint string, data []byte) (
	resp HttpResp, body []byte, err error) {

	client := &http.Client{}
	url := fmt.Sprintf("https://%s:9999/%v", p.pccIp, endPoint)
	req, _ := http.NewRequest(op, url, bytes.NewBuffer(data))
	req.Header.Add("Authorization", p.bearer)
	r, _ := client.Do(req)
	defer r.Body.Close()
	body, _ = ioutil.ReadAll(r.Body)

	var (
		rg       respGeneric
		dataJson []byte
	)
	if err = json.Unmarshal(body, &rg); err != nil {
		fmt.Printf("Unmarshalling Error:\n%v\n", string(body))
		return
	}

	if dataJson, err = json.Marshal(rg.Data); err != nil {
		fmt.Printf("Marshalling Error:\n%v\n", rg.Data)
		return
	}
	resp = HttpResp{
		Path:    rg.Path,
		Status:  rg.Status,
		Message: rg.Message,
		Error:   rg.Error,
		Data:    dataJson,
	}
	return
}

func (p PccClient) pccUserManagement(op string, endPoint string, data []byte) (
	body []byte, err error) {

	client := &http.Client{}
	url := fmt.Sprintf("https://%s:9999/%v", p.pccIp, endPoint)
	req, _ := http.NewRequest(op, url, bytes.NewBuffer(data))
	req.Header.Add("Authorization", p.bearer)
	req.Header.Add("Content-type", "application/json")
	r, _ := client.Do(req)
	defer r.Body.Close()
	body, _ = ioutil.ReadAll(r.Body)
	if r.StatusCode == 200 {
		return
	}

	var rg respGeneric
	if err = json.Unmarshal(body, &rg); err != nil {
		fmt.Printf("Unmarshalling Error:\n%v\n", string(body))
		return
	}
	err = fmt.Errorf("%v: %v", r.Status, rg.Error)
	return
}
