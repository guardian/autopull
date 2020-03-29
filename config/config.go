package config

import (
	"io/ioutil"
	"os"
	"gopkg.in/yaml.v2"
)

type Configuration struct {
	VaultDoorUri string `yaml:"vaultdoor_uri"`

}

func LoadConfig(path string) (*Configuration, error) {
	fp, openErr := os.OpenFile(path, os.O_RDONLY, 0)
	if openErr != nil {
		return nil, openErr
	}

	content, readErr := ioutil.ReadAll(fp)
	fp.Close()

	if readErr != nil {
		return nil, readErr
	}

	var config Configuration
	marshalErr := yaml.Unmarshal(content, &config)
	if marshalErr != nil {
		return nil, marshalErr
	}

	return &config, nil
}