package config

import (
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

var Version = "0.1.0"

var (
	instance *Config
	once     sync.Once
)

type Config struct {
	Token   string `yaml:"token"`
	BaseURL string `yaml:"base_url"`
}

func defaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mizbancloud", "config.yaml")
}

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{
			BaseURL: "https://auth.mizbancloud.com/api",
		}
		instance.Load()
	})
	return instance
}

func (c *Config) Load() error {
	path := defaultConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, c)
}

func (c *Config) Save() error {
	path := defaultConfigPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func (c *Config) SetToken(token string) error {
	c.Token = token
	return c.Save()
}

func (c *Config) SetBaseURL(url string) error {
	c.BaseURL = url
	return c.Save()
}

func (c *Config) IsLoggedIn() bool {
	return c.Token != ""
}

func (c *Config) Logout() error {
	c.Token = ""
	return c.Save()
}
