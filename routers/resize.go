package routers

import (
	"errors"
	"github.com/1000Delta/imageserver/models"
	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
	"image"
	"image/png"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func ImageResize(ctx *gin.Context) {
	id, exist := ctx.GetQuery("id")
	if !exist {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &RtnJsonBase{Msg: "缺失参数"})
		return
	}
	s, exist := ctx.GetQuery("size")
	if !exist {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &RtnJsonBase{Msg: "缺失参数"})
		return
	}
	interp := ctx.DefaultQuery("func", "nearestneighbor")
	// 参数检查
	idn, err := strconv.Atoi(id)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &RtnJsonBase{Msg: "id非数字"})
		return
	}
	w, h, err := parseSize(s)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &RtnJsonBase{Msg: "size无法解析"})
		return
	}
	// 加载图片数据
	imgInfo, err := models.GetImageByID(idn)
	if err != nil {
		if err == models.ErrRecordNotFound {
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}
		log.Printf("ImageResize get image info error: %v", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, &RtnJsonBase{Msg: "获取图片信息失败"})
		return
	}
	img, err := models.GetImageInStorage("." + imgInfo.Path)
	if err != nil {
		log.Printf("ImageResize get image info error: %v", err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, &RtnJsonBase{Msg: "获取图片失败"})
		return
	}
	var nImg image.Image
	switch interp {
	case "nearestneighbor":
		nImg = resize.Resize(uint(w), uint(h), img, resize.NearestNeighbor)
	case "bicubic":
		nImg = resize.Resize(uint(w), uint(h), img, resize.Bicubic)
	case "bilinear":
		nImg = resize.Resize(uint(w), uint(h), img, resize.Bilinear)
	case "lanczos2":
		nImg = resize.Resize(uint(w), uint(h), img, resize.Lanczos2)
	case "lanczos3":
		nImg = resize.Resize(uint(w), uint(h), img, resize.Lanczos3)
	case "mitchellnetravali":
		nImg = resize.Resize(uint(w), uint(h), img, resize.MitchellNetravali)
	default:
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &RtnJsonBase{Msg: "func无效"})
		return
	}
	err = png.Encode(ctx.Writer, nImg)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, &RtnJsonBase{Msg: "图片编码失败"})
	}
}

// 解析尺寸字符串，返回值为(width, height, error)
func parseSize(s string) (int, int, error) {
	size := strings.Split(s, "x")
	if len(size) != 2 {
		return 0, 0, errors.New("无法解析")
	}
	width, err := strconv.Atoi(size[0])
	if err != nil {
		return 0, 0, errors.New("参数非数字")
	}
	height, err := strconv.Atoi(size[1])
	if err != nil {
		return 0, 0, errors.New("参数非数字")
	}
	return width, height, nil
}
