package routers

import (
	"github.com/gin-gonic/gin"
)

func BaseRouteFunc(ctx *gin.Context) {
	ctx.String(200, "Welcome to imageserver")
}
