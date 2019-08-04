package main

import (
	"github.com/mgerb/mgdiscord/config"
	"github.com/mgerb/mgdiscord/connection"
)

func init() {
	config.Init()
}

func main() {
	connection.Start(config.Config.Token)
}
