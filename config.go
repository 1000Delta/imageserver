package main

import (
	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"log"
)

func configure() {
	data, err := ioutil.ReadFile("./config.toml")
	if err != nil {
		log.Fatalf("Config file error: %v", err.Error())
	}
	cfg, err := toml.LoadBytes(data)
	if err != nil {
		log.Fatalf("Parse config error: %v", err.Error())
	}

	if cfg.Has("mode") {
		mode := cfg.Get("mode").(string)
		switch mode {
		case "release":
			gin.SetMode(gin.ReleaseMode)
		case "debug":
			gin.SetMode(gin.DebugMode)
		case "tests":
			gin.SetMode(gin.TestMode)
		}
	}

	if cfg.Has("MaxMultipartMemory") {
		app.MaxMultipartMemory = cfg.Get("MaxMultipartMemory").(int64) << 20
	}

}
