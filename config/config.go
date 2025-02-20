package config

import (
	"os"

	"github.com/apieat/aigw/platform"
	"github.com/pkg/errors"
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
		return nil, errors.Wrap(err, "failed to parse yaml file")
	}

	err = config.Init()

	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) Init() error {
	err := c.Platform.Init()

	if err != nil {
		return errors.Wrap(err, "failed to init platform")
	}

	err = c.Api.Init()
	if err != nil {
		return errors.Wrap(err, "failed to init api")
	}
	return nil
}
