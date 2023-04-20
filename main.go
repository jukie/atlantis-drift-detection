package main

import (
	"flag"
	"log"
	"os"

	"github.com/jukie/atlantis-drift-detection/internal/config"
	"github.com/jukie/atlantis-drift-detection/internal/drift"
	"github.com/jukie/atlantis-drift-detection/internal/vcs"
)

func main() {
	var gitlabToken, githubToken string

	// Define flags
	flag.StringVar(&gitlabToken, "gitlab-token", os.Getenv("GITLAB_TOKEN"), "API token for Gitlab")
	flag.StringVar(&githubToken, "github-token", os.Getenv("GITHUB_TOKEN"), "API token for Github")
	flag.Parse()

	validateTokens(gitlabToken, githubToken)

	driftCfg, err := config.GetDriftCfg()
	if err != nil {
		log.Fatalln(err)
	}
	servers, err := config.LoadVcsConfig(driftCfg.ConfigPath)
	if err != nil {
		log.Fatalln(err)
	}

	processRepositories(servers, githubToken, gitlabToken, driftCfg)
}

func validateTokens(gitlabToken, githubToken string) {
	if gitlabToken == "" && githubToken == "" {
		log.Fatalln("Error: Both GitLab and GitHub tokens are not provided but at least one is required. Set GITLAB_TOKEN or GITHUB_TOKEN environment variables, or pass them using the --gitlab-token and/or --github-token flags.")
	}
}

func processRepositories(servers *config.VcsServers, githubToken, gitlabToken string, driftCfg config.DriftCfg) {
	if servers.GithubServer != nil {
		processGithubRepositories(servers.GithubServer, githubToken, driftCfg)
	}
	if servers.GitlabServer != nil {
		processGitlabRepositories(servers.GitlabServer, gitlabToken, driftCfg)
	}
}

func processGithubRepositories(ghServer *config.ServerCfg, githubToken string, driftCfg config.DriftCfg) {
	ghClient, err := vcs.NewGithubClient(ghServer.ApiEndpoint, githubToken)
	if err != nil {
		log.Fatalln("failed to setup github client")
	}
	for _, r := range ghServer.Repos {
		err := drift.Run(ghClient, r, driftCfg)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func processGitlabRepositories(glServer *config.ServerCfg, gitlabToken string, driftCfg config.DriftCfg) {
	glClient, err := vcs.NewGitlabClient(glServer.ApiEndpoint, gitlabToken)
	if err != nil {
		log.Fatalln("failed to setup gitlab client")
	}
	for _, r := range glServer.Repos {
		err := drift.Run(glClient, r, driftCfg)
		if err != nil {
			log.Fatalln(err)
		}
	}
}
