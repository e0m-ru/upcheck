package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Interval int          `yaml:"interval"`
	Sites    *SitesList   `yaml:"sites"`
	Alerts   *AlertConfig `yaml:"alerts"`
}

type SitesList []Site

type Site struct {
	URL     string `yaml:"url"`
	Timeout int    `yaml:"timeout"`
}

type AlertConfig struct {
	Enabled  bool         `yaml:"enabled"`
	Failures int          `yaml:"failures"`
	Cooldown int          `yaml:"cooldown"`
	Email    *EmailConfig `yaml:"email"`
}

type EmailConfig struct {
	To       string `yaml:"to"`
	SMTPHost string `yaml:"smtp_host"`
	SMTPPort string `yaml:"smtp_port"`
	SMTPUser string `yaml:"smtp_user"`
	SMTPPass string `yaml:"smtp_password"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	return &config, err
}
