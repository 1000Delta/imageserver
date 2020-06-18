package main

import (
	"github.com/1000Delta/imageserver/routers"
)

func route() {
	app.GET("/", routers.BaseRouteFunc)

	// image
	app.GET("/image", routers.ImageGet)
	app.POST("/image", routers.ImageUpload)
	app.DELETE("/image", routers.ImageRemove)
	app.PUT("/image", routers.ImageReplace)

	// static
	app.Static("/uploads", "./uploads")
}
