package main

import (
	"os"

	"github.com/platfornow/lash/cmd"
	"github.com/platfornow/lash/internal/log"
)

var (
	Version string
	Build   string
)

func main() {
	ctl := cmd.LandscapeShell(Version, Build)
	if err := ctl.Execute(); err != nil {
		log.GetInstance().Error(err.Error())
		os.Exit(1)
	}
}
