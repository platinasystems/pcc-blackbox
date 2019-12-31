package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type HttpResp struct {
	Path    string
	Status  int
	Message string
	Error   string
	Data    []byte
}

func pccGateway(op string, endPoint string, data []byte) (resp HttpResp, body []byte, err error) {
	client := &http.Client{}
	url := fmt.Sprintf("https://%s:9999/%v", Env.PccIp, endPoint)
	req, _ := http.NewRequest(op, url, bytes.NewBuffer(data))
	req.Header.Add("Authorization", Bearer)
	r, _ := client.Do(req)
	defer r.Body.Close()
	body, _ = ioutil.ReadAll(r.Body)
	type respGeneric struct {
		Path    string `json:"path"`
		Status  int    `json:"status"`
		Message string `json:"message"`
		Error   string `json:"error"`
		Data    interface{}
	}
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

func pccSecurity(op string, endPoint string, data []byte) (resp HttpResp, body []byte, err error) {
	client := &http.Client{}
	url := fmt.Sprintf("https://%s:9999/%v", Env.PccIp, endPoint)
	req, _ := http.NewRequest(op, url, bytes.NewBuffer(data))
	req.Header.Add("Authorization", Bearer)
	r, _ := client.Do(req)
	defer r.Body.Close()
	body, _ = ioutil.ReadAll(r.Body)
	//var (
	//	rg       []securityKey
	//	dataJson []byte
	//)

	//if err = json.Unmarshal(body, &rg); err != nil {
	//	fmt.Printf("Unmarshalling Error:\n%v\n", string(body))
	//	return
	//}
	//
	//
	//if dataJson, err = json.Marshal(rg); err != nil {
	//	fmt.Printf("Marshalling Error:\n%v\n", rg)
	//	return
	//}

	resp = HttpResp{
		Status: r.StatusCode,
		Data:   body,
	}
	return
}
