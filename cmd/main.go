package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/nfort/gopher-bot/internal/modules"
	"github.com/nfort/gopher-bot/internal/modules/config"

	"github.com/gin-gonic/gin"
)

const VERSION = "v1.2.1"

func main() {
	if err := config.InitConfig(); err != nil {
		log.Fatal(err)
	}

	portFlag := flag.Int("port", config.Config.Server.Port, "port on which the server listens")
	versionFlag := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	config.Config.Server.Port = *portFlag

	if !config.Config.Server.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.POST("/hook", func(c *gin.Context) {
		modules.HandlerWebHook(c)
	})

	if err := r.Run(getAddrStr()); err != nil {
		log.Fatal(err)
	}
}

func getAddrStr() string {
	host := config.Config.Server.Domain
	port := config.Config.Server.Port
	return fmt.Sprintf("%s:%d", host, port)
}
