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
	// 图像算法示例
	app.GET("/image/gaussFuzzy", routers.ImageGaussianBlur)
	app.GET("image/resize", routers.ImageResize)

	// static
	app.Static("/uploads", "./uploads")
}
