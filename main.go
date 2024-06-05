package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/yeahdongcn/topology/cmd"
)

func main() {
	log.SetLevel(log.DebugLevel)

	cmd.Execute()
}
