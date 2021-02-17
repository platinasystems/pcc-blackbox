// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pcc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/viper"
)

func LoadLogConfig(path string, format string) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Failed to load log configuration: %s", err)
	}
	storeToViper(content, format)
}

func storeToViper(content []byte, format string) {
	viper.SetConfigType(format)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	err := viper.ReadConfig(bytes.NewBuffer(content))
	if err != nil {
		fmt.Println(err)
	}
}
