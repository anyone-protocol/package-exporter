package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type PackageType int

const (
	PackageUnknown PackageType = iota
	PackageGitHubReleases
	PackageDockerHub
	PackageNginx
)

type Config struct {
	Fetchers FetcherConfigs `yaml:"fetchers"`
}

type FetcherConfigs struct {
	DockerHub       []DockerHubConfig       `yaml:"dockerhub_pulls"`
	GithubReleases  []GitHubReleasesConfig  `yaml:"github_releases"`
	NginxAccessLogs []NginxAccessLogsConfig `yaml:"nginx_access_log"`
}

type DockerHubConfig struct {
	Name  string `yaml:"name"`
	Owner string `yaml:"owner"`
	Repo  string `yaml:"repo"`
}

type GitHubReleasesConfig struct {
	Name         string `yaml:"name"`
	Owner        string `yaml:"owner"`
	Repo         string `yaml:"repo"`
	AssetsRegexp string `yaml:"assets_regexp"`
}

type NginxAccessLogsConfig struct {
	Name            string `yaml:"name"`
	AccessLogPath   string `yaml:"access_log_path"`
	AccessLogRegexp string `yaml:"access_log_regexp"`
}

// New creates new Config from config data
func New(data []byte) (*Config, error) {
	var cfg Config
	err := yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// FromFile reads filename and creates Config from it
func FromFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	cfg, err := New(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}

	return cfg, nil
}
