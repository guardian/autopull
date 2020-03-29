package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type Configuration struct {
	VaultDoorUri    string `yaml:"vaultdoor_uri"`
	DownloadThreads int    `yaml:"download_threads"`  //defaults to 5 if not specified
	QueueBufferSize int    `yaml:"queue_buffer_size"` //defaults to 10 if not specified
	AllowOverwrite  bool   `yaml:"allow_overwrite"`   //defaults to false
	DownloadPath    string `yaml:"download_path"`     //path to download to. Can be overridden on the commandline.
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
