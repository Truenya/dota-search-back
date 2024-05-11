package main

import (
	"os"

	"github.com/Truenya/dota-search-back/cache"
	"github.com/Truenya/dota-search-back/db"
	"github.com/Truenya/dota-search-back/miners"
	"github.com/Truenya/dota-search-back/net"
	log "github.com/sirupsen/logrus"
)

func InitLogs() {
	file, err := os.OpenFile("./vk-spoofer.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		panic(err)
	}

	log.SetOutput(file)
	log.SetLevel(log.TraceLevel)
	log.SetFormatter(&log.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: "2006/01/02 15:04:05"})
	log.Infoln("Backend started")
}

func main() {
	InitLogs()

	pg := db.Init("dotasearch")
	defer pg.Close()

	// collectors is not exactly needed to parse periodically
	// for better performance use it only for first time
	// cvk := collectors.NewVKCollector()
	// cvk.Collect()
	cache.Init()
	net.Init()

	c := miners.Controller{}

	c.Mine()
	defer c.Stop()

	net.RouteAll()
	log.Infoln("Backend finished")
}
