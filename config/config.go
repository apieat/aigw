package config

import (
	"os"

	"github.com/apieat/aigw/platform"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Api      ApiConfig         `yaml:"api"`
	Platform platform.AIConfig `yaml:"ai"`
}

func Init(fileName string) (cfg *Config, err error) {
	//read yaml file
	file, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	//parse yaml file to Config struct
	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	err = config.Platform.Init()

	if err != nil {
		return nil, err
	}

	err = config.Api.Init()
	if err != nil {
		return nil, err
	}

	return &config, nil
}
