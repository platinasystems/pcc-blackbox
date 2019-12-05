package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/platinasystems/tiles/pccserver/models"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	PUBLIC_KEY = "public_key"
	PRIVATE_KEY = "private_key"
	CERT = "certificate"

	UPLOAD = "upload"
	DESCRIBE = "describe"
	DELETE = "delete"
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


func BuildKeyManagerEndpoint(fileType string, label string, method string, id uint64) string{

	var relativePath string
	if method != DELETE {

		switch fileType {
		case PRIVATE_KEY:
			relativePath = "keys/upload/private"
			if method == DESCRIBE {
				relativePath = "keys/describe"
				return fmt.Sprintf("%s/%s/%v", KEYMANAGER_ENDPOINT, relativePath, label)
			}

		case PUBLIC_KEY:
			relativePath = "keys/upload/public"
			if method == DESCRIBE {
				relativePath = "keys/describe"
				return fmt.Sprintf("%s/%s/%v", KEYMANAGER_ENDPOINT, relativePath, label)
			}

		case CERT:

			relativePath = fmt.Sprintf( "certificates/%s", method)
			if method == DESCRIBE {
				if id > 0 {
					relativePath = fmt.Sprintf(relativePath + "/%v", id)
				}
				return fmt.Sprintf("%s/%s", KEYMANAGER_ENDPOINT, relativePath)
			}
		}
		return fmt.Sprintf("https://%s:9999/%s/%s/%v", Env.PccIp, KEYMANAGER_ENDPOINT, relativePath, label)

	} else {
		relativePath = "certificates"
		if fileType != CERT {
			relativePath = "keys"
		}
		return fmt.Sprintf("%s/%s/%v", KEYMANAGER_ENDPOINT, relativePath, id)
	}

}


func GetSecurityKey(alias string, keyType string) (securityKey, error){
	var securityKey securityKey
	endpoint := BuildKeyManagerEndpoint(keyType, alias, DESCRIBE, 0)
	resp, body, err := pccSecurity("GET", endpoint, nil)
	if err != nil {
		fmt.Printf("Get key %s failed\n%v\n", alias, string(body))
		return securityKey, err
	}
	err = json.Unmarshal(resp.Data, &securityKey)
	if err != nil {
		fmt.Printf("Umarshal key %s failed\n%v\n", alias, err)
	}
	return securityKey, err
}


func GetCertificate(alias string) ( certificate Certificate, err error){
	var certificates[] Certificate
	endpoint := BuildKeyManagerEndpoint(CERT, alias, DESCRIBE, 0)
	resp, body, err := pccSecurity("GET", endpoint, nil)
	if err != nil {
		fmt.Printf("Get certificates %s failed\n%v\n", alias, string(body))
		return certificate, err
	}
	err = json.Unmarshal(resp.Data, &certificates)
	if err == nil {
		for i := 0; i < len(certificates); i++ {
			if certificates[i].Alias == alias {
				certificate = certificates[i]
			}
		}
	}else {
		fmt.Printf("Unmarshal certificate %s failed\n%v\n", alias, err)
	}

	return certificate, err
}


func CreateFileAndUpload(fileName string, key string, fileType string)  error{

	var err error
	var f *os.File
	f, err = os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to create file:\n%v\n", err))
	}
	_, err = f.Write([]byte(key))
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to write on disk:\n%v\n", err))
	}
	f.Close()

	filePath := fmt.Sprintf("./%s",fileName)

	exist, id, err := checkAliasExist(fileName, fileType)

	if exist {
		removeKeyCert(id, fileType)
	}

	url := BuildKeyManagerEndpoint(fileType, fileName, UPLOAD, 0)
	err = UpdateFile(filePath, url)
	if err != nil {
		return err
	}

	return err
}

func removeKeyCert(id uint64, fileType string) error{
	url := BuildKeyManagerEndpoint(fileType, "", DELETE, id)
	_, body, err := pccSecurity("DELETE", url, nil)
	if err != nil {
		fmt.Printf("DELETE %s failed\n%v\n", fileType, string(body))
	}
	return err
}

func checkAliasExist(alias string, fileType string) (exist bool, id uint64, err error){
	url := BuildKeyManagerEndpoint(fileType, alias, DESCRIBE, 0)

	resp, body, err := pccSecurity("GET", url, nil)
	if err != nil {
		fmt.Printf("%s %s not found \n%v\n", fileType, alias, string(body))
		return false, 0, err
	} else {
		if fileType != CERT {
			var secKey securityKey
			err = json.Unmarshal(resp.Data, &secKey)
			if err == nil {
				exist = true
				id = secKey.Id
			}
		}else {
			var certificates [] Certificate
			err = json.Unmarshal(resp.Data, &certificates)
			if err == nil {
				for i := 0; i < len(certificates); i++ {
					if certificates[i].Alias == alias {
						exist = true
						id = certificates[i].Id
					}
				}
			}
		}
	}
	return
}


func IsInvader(node *models.NodeWithKubernetes) bool{
	for i := 0; i < len(Env.Invaders); i++ {
		if Env.Invaders[i].HostIp == node.Host {
			return true
		}
	}
	return false
}

func ConvertToMillis(startTime time.Time) uint64 {
	return uint64(startTime.UnixNano()) / uint64(time.Millisecond)
}