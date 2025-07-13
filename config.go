package main

import (
	"os"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Port int `yaml:"port"`
}

func loadConfig() (Config, error) {
	var config Config

	file, err := os.ReadFile("config.yaml")
	if err != nil {
		return config, err
	}

	if err := yaml.Unmarshal(file, &config); err != nil {
		return config, err
	}

	return config, nil
}
