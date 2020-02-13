package main

import (
	"fmt"
	"os"
)

const (
	PUBLIC_KEY  = "public_key"
	PRIVATE_KEY = "private_key"
	CERT        = "certificate"

	UPLOAD   = "upload"
	DESCRIBE = "describe"
	DELETE   = "delete"
)

func CreateFileAndUpload(fileName string, key string, fileType string) (err error) {

	var f *os.File
	f, err = os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		err = fmt.Errorf("Unable to create file:\n%v\n", err)
		return
	}
	defer f.Close()
	_, err = f.Write([]byte(key))
	if err != nil {
		err = fmt.Errorf("Unable to write on disk:\n%v\n", err)
		return
	}

	filePath := fmt.Sprintf("./%s", fileName)

	// check if exist and delete if so

	err = Pcc.UploadCert(filePath, "foo", "")
	if err != nil {
		return
	}

	return
}
