package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type nodeAddReq struct {
	Id          uint64
	Name        string
	Host        string
	SN          string
	Model       string
	Tags        []string
	Site_Id     uint64
	ClusterId   uint64
	Owner       uint64
	Roles       []uint64
	Bmc         string
	BmcUser     string
	BmcUsers    []string
	BmcPassword string
	AdminUser   string
	SSHKeys     []uint64
	Console     string
	Managed     bool
	Standby     bool
}

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
		return
	}

	if dataJson, err = json.Marshal(rg.Data); err != nil {
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
