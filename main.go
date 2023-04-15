package main

import (
	"flag"
	"log"
	"os"

	"github.com/jukie/atlantis-drift-detection/config"
	"github.com/jukie/atlantis-drift-detection/pkg/drift"
	"github.com/jukie/atlantis-drift-detection/pkg/vcs"
)

func main() {

	var gitlabToken, githubToken string

	// Define flags
	flag.StringVar(&gitlabToken, "gitlab-token", os.Getenv("GITLAB_TOKEN"), "API token for Gitlab")
	flag.StringVar(&githubToken, "github-token", os.Getenv("GITHUB_TOKEN"), "API token for Github")
	if gitlabToken == "" && githubToken == "" {
		log.Fatalln("Error: Both GitLab and GitHub tokens are not provided but at least one is required. Set GITLAB_TOKEN or GITHUB_TOKEN environment variables, or pass them using the --gitlab-token and/or --github-token flags.")
	}

	driftCfg, err := config.GetDriftCfg()
	if err != nil {
		log.Fatalln(err)
	}
	servers, err := config.LoadVcsConfig(driftCfg.ConfigPath)
	if err != nil {
		log.Fatalln(err)
	}
	ghClient, err := vcs.NewGithubClient(servers.GithubServer.ApiEndpoint, githubToken)
	glClient, err := vcs.NewGitlabClient(servers.GitlabServer.ApiEndpoint, gitlabToken)
	for _, r := range servers.GitlabServer.Repos {
		err := drift.Run(glClient, r, driftCfg)
		if err != nil {
			log.Fatalln(err)
		}
	}
	for _, r := range servers.GithubServer.Repos {
		err := drift.Run(ghClient, r, driftCfg)
		if err != nil {
			log.Fatalln(err)
		}
	}
}
