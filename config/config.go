package config

import (
	"fmt"
	"os"

	"github.com/jukie/atlantis-drift-detection/pkg/vcs"
	"gopkg.in/yaml.v3"
)

type DriftCfg struct {
	AtlantisUrl   string
	AtlantisToken string
	ConfigPath    string
}
type Repo struct {
	Ref  string
	Name string
}

type ServerCfg struct {
	ApiEndpoint string            `yaml:"apiEndpoint"`
	Client      *vcs.GitlabClient `yaml:"-"`
	Token       string            `yaml:"token"`
	Repos       []Repo            `yaml:"repos"`
}

type VcsServers struct {
	GithubServers ServerCfg `yaml:"github"`
	GitlabServers ServerCfg `yaml:"gitlab"`
}

func GetDriftCfg() (DriftCfg, error) {
	var d DriftCfg

	var url, token, configPath string
	var ok bool

	if url, ok = os.LookupEnv("ATLANTIS_URL"); !ok {
		return d, fmt.Errorf("ATLANTIS_URL environment variable is required but not set")
	}
	d.AtlantisUrl = url

	if token, ok = os.LookupEnv("ATLANTIS_TOKEN"); !ok {
		return d, fmt.Errorf("ATLANTIS_TOKEN environment variable is required but not set")
	}
	d.AtlantisToken = token

	if configPath, ok = os.LookupEnv("CONFIG_PATH"); !ok {
		return d, fmt.Errorf("CONFIG_PATH environment variable is required but not set")
	}
	d.ConfigPath = configPath

	return d, nil
}

func LoadVcsConfig(repoCfgPath string) (VcsServers, error) {
	var cfg VcsServers
	if fileExists(repoCfgPath) {
		f, err := os.ReadFile(repoCfgPath)
		if err != nil {
			return cfg, err
		}
		err = yaml.Unmarshal(f, &cfg)
		if err != nil {
			return cfg, err
		}
		return cfg, nil
	}
	return cfg, fmt.Errorf("could not find config file")
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
