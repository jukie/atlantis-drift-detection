package config_test

import (
	"os"
	"testing"

	"github.com/jukie/atlantis-drift-detection/config"
	"github.com/stretchr/testify/assert"
)

func TestGetDriftCfg(t *testing.T) {
	os.Setenv("ATLANTIS_URL", "http://example.com")
	os.Setenv("ATLANTIS_TOKEN", "token")
	os.Setenv("CONFIG_PATH", "/path/to/config.yaml")
	defer os.Clearenv()

	expectedCfg := config.DriftCfg{
		AtlantisUrl:   "http://example.com",
		AtlantisToken: "token",
		ConfigPath:    "/path/to/config.yaml",
	}

	cfg, err := config.GetDriftCfg()
	assert.NoError(t, err)
	assert.Equal(t, expectedCfg, cfg)
}

func TestGetDriftCfgMissingEnvVar(t *testing.T) {
	os.Unsetenv("ATLANTIS_URL")
	os.Unsetenv("ATLANTIS_TOKEN")
	os.Unsetenv("CONFIG_PATH")
	defer os.Clearenv()

	_, err := config.GetDriftCfg()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ATLANTIS_URL environment variable is required but not set")
}

func TestLoadVcsConfig(t *testing.T) {
	cfgYAML := `github:
  apiEndpoint: https://api.github.com
  repos:
  - ref: master
    name: repo1
gitlab:
  apiEndpoint: https://gitlab.com/api/v4
  repos:
  - ref: main
    name: repo2
`
	tmpfile, err := os.CreateTemp("", "config")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	err = os.WriteFile(tmpfile.Name(), []byte(cfgYAML), 0644)
	assert.NoError(t, err)

	expectedCfg := config.VcsServers{
		GithubServer: config.ServerCfg{
			ApiEndpoint: "https://api.github.com",
			Repos: []config.Repo{
				{Ref: "master", Name: "repo1"},
			},
		},
		GitlabServer: config.ServerCfg{
			ApiEndpoint: "https://gitlab.com/api/v4",
			Repos: []config.Repo{
				{Ref: "main", Name: "repo2"},
			},
		},
	}

	cfg, err := config.LoadVcsConfig(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, expectedCfg, cfg)
}

func TestLoadVcsConfigMissingFile(t *testing.T) {
	cfgPath := "/path/to/nonexistent.yaml"

	_, err := config.LoadVcsConfig(cfgPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not find config file")
}
