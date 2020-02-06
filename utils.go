package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/platinasystems/tiles/pccserver/models"
)

func UpdateFile(filePath string, url string) (err error) {
	file, err := os.Open(filePath)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
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
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Add("Authorization", Bearer)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r, err := client.Do(req)
	defer r.Body.Close()
	return err
}

func IsInvader(node *models.NodeWithKubernetes) bool {
	for i := 0; i < len(Env.Invaders); i++ {
		if Env.Invaders[i].HostIp == node.Host {
			return true
		}
	}
	return false
}

func IsOnline(node *models.NodeWithKubernetes) bool {
	if node.NodeAvailabilityStatus.ConnectionStatus == "online" {
		return true
	}
	return false
}

func ConvertToMillis(startTime time.Time) uint64 {
	return uint64(startTime.UnixNano()) / uint64(time.Millisecond)
}

func IsAppInstalled(nodeId uint64, appId string) (isInstalled bool) {

	endpoint := fmt.Sprintf("pccserver/node/%v/apps", nodeId)
	var (
		body []byte
		resp HttpResp
		err  error
		apps []models.ProvisionedApp
	)

	if resp, body, err = pccGateway("GET", endpoint, nil); err != nil {
		fmt.Printf("%v\n%v\n", string(body), err)
		return
	}
	if resp.Status != 200 {
		fmt.Printf("Get node apps failed\n%v\n", string(body))
		return
	}
	if err := json.Unmarshal(resp.Data, &apps); err != nil {
		fmt.Printf("%v\n%v\n", string(resp.Data), err)
		return
	}

	for i := range apps {
		if apps[i].ID == appId && apps[i].Local.Installed {
			isInstalled = true
			return
		}
	}
	return
}
