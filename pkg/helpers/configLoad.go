package helpers

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Tube struct {
	URL      string `yaml:"url"`
	TubeName string `yaml:"queue"`
}

type Config struct {
	Port             string `yaml:"port"`
	Tube             `yaml:"tube"`
	Tasks            map[string]string `yaml:"tasks"`
	WorkersCount     int               `yaml:"workers_count"`
	MaxRetryAtts     int               `yaml:"max_retry_attempts"`
	InitialTimeRetry int               `yaml:"initial_retry"`
}

func LoadConf(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf := Config{}
	if err := yaml.Unmarshal(file, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}
