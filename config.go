package main

import (
	"os"

	yaml "gopkg.in/yaml.v3"
)

type AppConfig struct {
	APIKey string `yaml:"apikey"`

	Webhooks WebhooksSettings    `yaml:"webhooks"`
	Submit   []SubmitterSettings `yaml:"submit"`
}

type WebhooksSettings struct {
	Enabled bool   `yaml:"enabled"`
	Listen  string `yaml:"listen"`
	Reports struct {
		Alerted   ReportOutput `yaml:"alerted"`
		Submitted ReportOutput `yaml:"submitted"`
	} `yaml:"reports"`
}

type SubmitterSettings struct {
	File     string `yaml:"file"`
	Enabled  bool   `yaml:"enabled"`
	Settings struct {
		Access string            `yaml:"access"`
		Tags   []string          `yaml:"tags"`
		Meta   map[string]string `yaml:"meta"`
	} `yaml:"settings"`
	Output ReportOutput `yaml:"output"`
}

type ReportOutput struct {
	Enabled     bool   `yaml:"enabled"`
	Path        string `yaml:"path"`
	Report      bool   `yaml:"report"`
	Screenshot  bool   `yaml:"screenshot"`
	DomainGraph bool   `yaml:"domain_graph"`
}

func LoadConfig(filename string) (*AppConfig, error) {
	var cfg AppConfig

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
