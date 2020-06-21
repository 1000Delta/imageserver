package routers

import (
	"github.com/1000Delta/imageserver/models"
	"github.com/1000Delta/imageserver/utils"
	"github.com/gin-gonic/gin"
	png2 "image/png"
	"log"
	"net/http"
	"strconv"
)

func ImageGaussianBlur(ctx *gin.Context) {
	id, exist := ctx.GetQuery("id")
	if !exist {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &RtnJsonBase{Msg: "缺失图片id"})
		return
	}
	r := ctx.DefaultQuery("radius", "5")
	s := ctx.DefaultQuery("sigma", "5")
	// 参数解析
	idn, err := strconv.Atoi(id)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &RtnJsonBase{Msg: "id非数字"})
		return
	}
	radius, err := strconv.Atoi(r)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &RtnJsonBase{Msg: "radius非数字"})
		return
	}
	sigma, err := strconv.ParseFloat(s, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &RtnJsonBase{Msg: "sigma非数字"})
		return
	}
	imgInfo, err := models.GetImageByID(idn)
	if err != nil {
		log.Printf("ImageGaussianBlur get image info error: %v", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, &RtnJsonBase{Msg: "读取图片信息失败"})
		return
	}
	img, err := models.GetImageInStorage("." + imgInfo.Path)
	if err != nil {
		log.Printf("ImageGaussianBlur get image in storage error: %v", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, &RtnJsonBase{Msg: "读取图片失败"})
		return
	}
	nImg := utils.GaussianBlur(img, radius, 1, sigma)
	err = png2.Encode(ctx.Writer, nImg)
	if err != nil {
		log.Printf("ImageGaussianBlur encode png error: %v", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, &RtnJsonBase{Msg: "图片编码失败"})
		return
	}
}
