package main

import (
	"github.com/koh-sh/codebuild-multirunner/cmd"
	_ "github.com/koh-sh/codebuild-multirunner/cmd/cwlog"
	_ "github.com/koh-sh/codebuild-multirunner/cmd/dump"
	_ "github.com/koh-sh/codebuild-multirunner/cmd/run"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersionInfo(version, commit, date)
	cmd.Execute()
}
