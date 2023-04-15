package config_test

import (
	"os"
	"testing"

	"github.com/jukie/atlantis-drift-detection/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setEnv(t *testing.T, key, value string) {
	t.Helper()
	require.NoError(t, os.Setenv(key, value))
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	require.NoError(t, os.Unsetenv(key))
}

func TestGetDriftCfg(t *testing.T) {
	setEnv(t, "ATLANTIS_URL", "https://example.com")
	setEnv(t, "ATLANTIS_TOKEN", "example_token")
	setEnv(t, "CONFIG_PATH", "example_config_path")
	defer unsetEnv(t, "ATLANTIS_URL")
	defer unsetEnv(t, "ATLANTIS_TOKEN")
	defer unsetEnv(t, "CONFIG_PATH")

	cfg, err := config.GetDriftCfg()
	require.NoError(t, err)

	assert.Equal(t, "https://example.com", cfg.AtlantisUrl)
	assert.Equal(t, "example_token", cfg.AtlantisToken)
	assert.Equal(t, "example_config_path", cfg.ConfigPath)
}

func TestLoadVcsConfig(t *testing.T) {
	cfgContent := `
github:
  apiEndpoint: https://api.github.com
  token: github_token
  repos:
    - ref: main
      name: jukie/repo1
    - ref: master
      name: jukie/repo2
gitlab:
  apiEndpoint: https://gitlab.com/api/v4
  token: gitlab_token
  repos:
    - ref: main
      name: jukie/repo3
`
	cfgFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(cfgFile.Name())

	_, err = cfgFile.WriteString(cfgContent)
	require.NoError(t, err)

	cfg, err := config.LoadVcsConfig(cfgFile.Name())
	require.NoError(t, err)

	assert.Equal(t, "https://api.github.com", cfg.GithubServers.ApiEndpoint)
	assert.Equal(t, "github_token", cfg.GithubServers.Token)
	assert.Equal(t, "https://gitlab.com/api/v4", cfg.GitlabServers.ApiEndpoint)
	assert.Equal(t, "gitlab_token", cfg.GitlabServers.Token)

	assert.Len(t, cfg.GithubServers.Repos, 2)
	assert.Len(t, cfg.GitlabServers.Repos, 1)

	assert.Equal(t, "jukie/repo1", cfg.GithubServers.Repos[0].Name)
	assert.Equal(t, "main", cfg.GithubServers.Repos[0].Ref)
	assert.Equal(t, "jukie/repo2", cfg.GithubServers.Repos[1].Name)
	assert.Equal(t, "master", cfg.GithubServers.Repos[1].Ref)
	assert.Equal(t, "jukie/repo3", cfg.GitlabServers.Repos[0].Name)
	assert.Equal(t, "main", cfg.GitlabServers.Repos[0].Ref)
}
