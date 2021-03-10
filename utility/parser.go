package utility

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type Subtests []string

type CustomTests struct {
	TestList map[string]Subtests `yaml:"tests"`
}

func GetCustomTests(fileName string) (tests CustomTests, err error) {
	if _, err = os.Stat(fileName); err == nil {
		var content []byte
		content, err = ioutil.ReadFile(fileName)

		if err != nil {
			err = errors.New(fmt.Sprintf("Failed to load the test list: %s", err))
			return
		}
		err = yaml.Unmarshal(content, &tests)
		if err != nil {
			err = errors.New(fmt.Sprintf("Failed to unmarshal the test list: %s", err))
			return
		}
	} else if os.IsNotExist(err) {
		err = errors.New(fmt.Sprintf("No file found with name %s", fileName))
	}
	return
}
