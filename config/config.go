package config

import (
	"errors"
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
	NoWait          bool   `yaml:"immediate_exit"`    //set to False on windows so you can see the result before the window shuts
}

func LoadConfig(path string) (conf *Configuration, err error) {
	defer func() {
		if r := recover(); r != nil {
			maybeError, isError := r.(error)
			if isError {
				conf = nil
				err = maybeError
			}
			maybeString, isString := r.(string)
			if isString {
				conf = nil
				err = errors.New(maybeString)
			}
			err = errors.New("unrecoverable error loading configuration")
		}
	}()

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
