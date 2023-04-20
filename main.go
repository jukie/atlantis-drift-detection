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
	executeDriftCheck(servers, githubToken, gitlabToken, driftCfg)
}

func validateTokens(gitlabToken, githubToken string) {
	if gitlabToken == "" && githubToken == "" {
		log.Fatalln("Error: Both GitLab and GitHub tokens are not provided but at least one is required. Set GITLAB_TOKEN or GITHUB_TOKEN environment variables, or pass them using the --gitlab-token and/or --github-token flags.")
	}
}

func executeDriftCheck(servers *config.VcsServers, githubToken, gitlabToken string, driftCfg config.DriftCfg) {
	if servers.GithubServer != nil {
		ghClient, err := vcs.NewGithubClient(servers.GithubServer.ApiEndpoint, githubToken)
		if err != nil {
			log.Fatalln("failed to setup github client")
		}
		driftRunner(ghClient, servers.GithubServer.Repos, driftCfg)
	}
	if servers.GitlabServer != nil {
		glClient, err := vcs.NewGitlabClient(servers.GitlabServer.ApiEndpoint, gitlabToken)
		if err != nil {
			log.Fatalln("failed to setup gitlab client")
		}
		driftRunner(glClient, servers.GitlabServer.Repos, driftCfg)
	}
}

func driftRunner(client vcs.Client, repos []config.Repo, driftCfg config.DriftCfg) {
	for _, r := range repos {
		err := drift.Run(client, r, driftCfg)
		if err != nil {
			log.Fatalln(err)
		}
	}
}
