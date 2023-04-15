package main

import (
	"log"

	"github.com/jukie/atlantis-drift-detection/config"
	"github.com/jukie/atlantis-drift-detection/pkg/drift"
)

func main() {
	driftCfg, err := config.GetDriftCfg()
	if err != nil {
		log.Fatalln(err)
	}
	servers, err := config.LoadVcsConfig(driftCfg.ConfigPath)
	if err != nil {
		log.Fatalln(err)
	}
	for _, r := range servers.GitlabServers.Repos {
		err := drift.Run(servers.GitlabServers.Client, r, driftCfg)
		if err != nil {
			log.Fatalln(err)
		}
	}
	for _, r := range servers.GithubServers.Repos {
		err := drift.Run(servers.GithubServers.Client, r, driftCfg)
		if err != nil {
			log.Fatalln(err)
		}
	}
}
