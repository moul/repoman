package main

import (
	"os"
	"path/filepath"

	hub "github.com/github/hub/v2/github"
)

func initMoulBotEnv() {
	if os.Getenv("REPOMAN_INITED") == "true" {
		return
	}
	os.Setenv("REPOMAN_INITED", "true")
	os.Setenv("GIT_AUTHOR_NAME", "moul-bot")
	os.Setenv("GIT_COMMITTER_NAME", "moul-bot")
	os.Setenv("GIT_AUTHOR_EMAIL", "bot@moul.io")
	os.Setenv("GIT_COMMITTER_EMAIL", "bot@moul.io")
	os.Setenv("HUB_CONFIG", filepath.Join(os.Getenv("HOME"), ".config", "hub-moul-bot"))
	config := hub.CurrentConfig()
	os.Setenv("GITHUB_TOKEN", config.Hosts[0].AccessToken)
}
