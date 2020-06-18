package main

import (
	"github.com/gin-gonic/gin"
	"log"
)

var app = gin.Default()

func main() {
	err := app.Run(":8080")
	if err != nil {
		log.Fatal(err.Error())
	}
}

func init() {
	configure()
	route()
}
