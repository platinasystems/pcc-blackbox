// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"fmt"
	"strconv"
)

const (
	PUBLIC_KEY  = "public"
	PRIVATE_KEY = "private"
	CERT        = "certificate"
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
	HasPrivate  bool   `json:"hasPrivate"`
	HasPublic   bool   `json:"hasPublic"`
}

func (p *PccClient) checkKeyType(fileType string) (err error) {
	if fileType != PRIVATE_KEY && fileType != PUBLIC_KEY {
		err = fmt.Errorf("Invalid security key type [%v]\n", fileType)
	}
	return
}

// Upload a key to the KM
func (p *PccClient) UploadKey(filePath string, label string, fileType string, description string) (key SecurityKey, err error) {
	if err = p.checkKeyType(fileType); err == nil {
		endPoint := fmt.Sprintf("key-manager/keys/upload/%s/%s", fileType, label)
		m := map[string]string{"description": description}
		err = p.PutFile(endPoint, filePath, m, &key)
	}
	return
}

// Download a key from the KM
func (p *PccClient) DownloadKey(id interface{}, fileType string) (content string, err error) {
	if err = p.checkKeyType(fileType); err == nil {
		endPoint := fmt.Sprintf("key-manager/keys/download/%s/%v", fileType, id)
		content, err = p.GetFile(endPoint)
	}
	return
}

func (p *PccClient) DeleteKey(id interface{}) (err error) { // THe KM accepts the deletion by ID or alias
	endpoint := fmt.Sprintf("key-manager/keys/%v", id)
	err = p.Delete(endpoint, nil, nil)
	return
}

func (p *PccClient) GetSecurityKeys() (secKeys []SecurityKey, err error) {
	endpoint := fmt.Sprintf("key-manager/keys/describe")
	err = p.Get(endpoint, &secKeys)
	return
}

func (p *PccClient) GetSecurityKey(alias string) (secKey SecurityKey, err error) {
	endpoint := fmt.Sprintf("key-manager/keys/describe/%v", alias)
	err = p.Get(endpoint, &secKey)
	return
}

// It is not returning a valid key
func (p *PccClient) UpdateSecurityKey(secKey SecurityKey) (err error) {
	err = p.Post("key-manager/keys/update", &secKey, &secKey)
	return
}

func (p *PccClient) FindSecurityKey(alias string) (exist bool, secKey SecurityKey, err error) {
	var secKeys []SecurityKey
	exist = false
	if secKeys, err = p.GetSecurityKeys(); err == nil {
		for _, k := range secKeys {
			if k.Alias == alias {
				secKey = k
				exist = true
				break
			}
		}
	}
	return
}

// It is not returning a valid certificate
func (p *PccClient) UploadCert(filePath string, label string, description string, keyId uint64) (certificate Certificate, err error) {
	endPoint := fmt.Sprintf("key-manager/certificates/upload/%s", label)
	m := map[string]string{"description": description}
	if keyId > 0 {
		m["keyId"] = strconv.FormatUint(keyId, 10)
	}
	err = p.PutFile(endPoint, filePath, m, &certificate)
	return
}

func (p *PccClient) UploadCertPrivatePublic(filePath string, keyPath string, alias string, description string) (certificate Certificate, err error) {
	endPoint := fmt.Sprintf("key-manager/certificates/upload/%s", alias)
	m := map[string]string{"description": description}
	fileMap := map[string]string{"file": filePath, "key": keyPath}
	err = p.PutFiles("POST", endPoint, fileMap, m, &certificate)
	return
}

// Download a certificate from the KM
func (p *PccClient) DownloadCertificate(id uint64) (content string, err error) {
	endPoint := fmt.Sprintf("key-manager/certificates/%d", id)
	content, err = p.GetFile(endPoint)
	return
}

func (p *PccClient) FindCertificate(alias string) (exist bool, certificate Certificate, err error) {
	var certificates []Certificate
	if certificates, err = p.GetCertificates(); err == nil {
		for i := range certificates {
			if certificates[i].Alias == alias {
				exist = true
				certificate = certificates[i]
				return
			}
		}
	}
	return
}

func (p *PccClient) GetCertificates() (certificates []Certificate, err error) {
	endpoint := fmt.Sprintf("key-manager/certificates/describe")
	err = p.Get(endpoint, &certificates)
	return
}

func (p *PccClient) GetCertificate(id uint64) (certificate Certificate, err error) {
	endpoint := fmt.Sprintf("key-manager/certificates/describe/%v", id)
	err = p.Get(endpoint, &certificate)
	return
}

func (p *PccClient) DeleteCertificate(id uint64) (err error) {
	endpoint := fmt.Sprintf("key-manager/certificates/%v", id)
	err = p.Delete(endpoint, nil, nil)
	return
}

func (p *PccClient) CheckKeyLabelExists(label string) (exists bool, err error) {
	var secKeys []SecurityKey

	if secKeys, err = p.GetSecurityKeys(); err == nil {
		for i := 0; i < len(secKeys); i++ {
			if secKeys[i].Alias == label {
				exists = true
			}
		}
	}
	return
}
