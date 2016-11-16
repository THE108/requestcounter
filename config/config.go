package config

import (
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	defaultListenHost       = "0.0.0.0"
	defaultListenPort       = 8080
	defaultIntervalCount    = 100
	defaultIntervalDuration = 600 * time.Millisecond
	defaultFilename         = "/tmp/requestcounter.dat"
	defaultPersistDuration  = 5 * time.Second
)

type Config struct {
	Host             string        `yaml:"host"`
	Port             int           `yaml:"port"`
	IntervalCount    int           `yaml:"interval-count"`
	IntervalDuration time.Duration `yaml:"interval-duration"`
	Persistent       bool          `yaml:"persistent"`
	Filename         string        `yaml:"filename"`
	PersistDuration  time.Duration `yaml:"persist-duration"`
}

func LoadConfigFromFile() (*Config, error) {
	cfg := &Config{}

	if filename := getConfigFilenameFromConsoleArgs(); filename != "" {
		if err := readAndUnmarshal(filename, cfg); err != nil {
			return nil, err
		}
	}

	cfg.setDefaults()

	return cfg, nil
}

func readAndUnmarshal(filename string, cfg *Config) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) setDefaults() {
	if cfg.Host == "" {
		cfg.Host = defaultListenHost
	}

	if cfg.Port == 0 {
		cfg.Port = defaultListenPort
	}

	if cfg.IntervalCount == 0 {
		cfg.IntervalCount = defaultIntervalCount
	}

	if cfg.IntervalDuration == 0 {
		cfg.IntervalDuration = defaultIntervalDuration
	}

	if cfg.Filename == "" {
		cfg.Filename = defaultFilename
	}

	if cfg.PersistDuration == 0 {
		cfg.PersistDuration = defaultPersistDuration
	}
}
