// manage.go
// 存放图片管理相关的路由逻辑
package routers

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/1000Delta/imageserver/models"
	"github.com/gin-gonic/gin"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	// TODO 修改路径为配置引入
	uploadPath = "./uploads/tmp/"
)

// 图片列表返回数据
type RtnJsonImageList struct {
	RtnJsonBase
	Total int             `json:"total"`
	Data  []*models.Image `json:"data"`
}

// 图片返回数据
type RtnJsonImage struct {
	RtnJsonBase
	Data *models.Image `json:"data"`
}

// 查询图片 优先级：id > name > 列表
func ImageGet(ctx *gin.Context) {
	if id, exist := ctx.GetQuery("id"); exist {
		idn, err := strconv.Atoi(id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, &RtnJsonBase{Msg: "id参数无效"})
			return
		}
		img, err := models.GetImageByID(idn)
		if err != nil {
			if err == models.ErrRecordNotFound {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}
			log.Printf("Query image by id error: %v", err.Error())
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, &RtnJsonBase{"查询图片信息失败"})
			return
		}
		ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s", img.Path))
		return
	}
	if name, exist := ctx.GetQuery("name"); exist {
		img, err := models.GetImageByName(name)
		if err != nil {
			if err == models.ErrRecordNotFound {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}
			log.Printf("Query image by name error: %v", err.Error())
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, &RtnJsonBase{"查询图片信息失败"})
			return
		}
		ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s", img.Path))
		return
	}
	images, err := models.GetImageList()
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, &RtnJsonBase{"查询图片列表失败"})
		return
	}
	ctx.JSON(http.StatusOK, &RtnJsonImageList{
		Data:  images,
		Total: len(images),
	})
}

func ImageUpload(ctx *gin.Context) {
	fh, err := ctx.FormFile("image")
	if err != nil {
		log.Printf("Upload error: %v\n", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &RtnJsonBase{Msg: "上传出错"})
		return
	}
	ext := filepath.Ext(fh.Filename)
	// 文件后缀检查
	if !isImageExtValid(ext) {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &RtnJsonBase{Msg: "文件后缀类型不支持"})
		return
	}
	// TODO 文件类型检查
	hash := genImageHash(fh.Filename)
	fPath := uploadPath + hash + ext
	err = ctx.SaveUploadedFile(fh, fPath)
	if err != nil {
		log.Printf("Save file error: %v\n", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &RtnJsonBase{Msg: "服务器保存图片失败"})
		return
	}

	nImg := &models.Image{
		Name: fh.Filename,
		Ext:  ext,
		Path: fPath[1:], // 存储为绝对路径
	}
	// 保存图片信息到数据库
	err = models.SaveNewImage(nImg)

	if err != nil {
		log.Printf("Save Image Error: %v", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &RtnJsonBase{Msg: "记录图片信息失败"})
		// 失败后删除图片
		err := os.Remove(fPath)
		if err != nil {
			log.Printf("Remove invalid image error: %v", err.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, &RtnJsonImage{
		RtnJsonBase: RtnJsonBase{Msg: "ok"},
		Data:        nImg,
	})
}

func ImageRemove(ctx *gin.Context) {
	if id, exist := ctx.GetQuery("id"); exist {
		idn, err := strconv.Atoi(id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, &RtnJsonBase{Msg: "id无效"})
			return
		}
		if err := models.RemoveImageByID(idn); err != nil {
			log.Printf("Remove image by id error: %v", err.Error())
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, &RtnJsonBase{Msg: "删除失败"})
			return
		}
		ctx.JSON(http.StatusOK, &RtnJsonBase{Msg: "ok"})
		return
	}
	if name, exist := ctx.GetQuery("name"); exist {
		if err := models.RemoveImageByName(name); err != nil {
			log.Printf("Remove image by name error: %v", err.Error())
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, &RtnJsonBase{Msg: "删除失败"})
			return
		}
		ctx.JSON(http.StatusOK, &RtnJsonBase{Msg: "ok"})
		return
	}
	ctx.AbortWithStatusJSON(http.StatusBadRequest, &RtnJsonBase{Msg: "缺失参数"})
}

func ImageReplace(ctx *gin.Context) {
	if name, exist := ctx.GetQuery("name"); exist {
		fh, err := ctx.FormFile("image")
		if err != nil {
			log.Printf("Image replace get file error: %v", err.Error())
			ctx.AbortWithStatusJSON(http.StatusBadRequest, &RtnJsonBase{Msg: "上传出错"})
			return
		}
		rand.Seed(time.Now().Unix())
		ext := filepath.Ext(fh.Filename)
		if !isImageExtValid(ext) {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, &RtnJsonBase{Msg: "文件类型不支持"})
			return
		}
		hash := genImageHash(fh.Filename)
		fPath := uploadPath + hash + ext
		// 更改记录并检查文件名是否有效
		if err := models.ReplaceImageByName(name, fPath[1:]); err != nil {
			if err == models.ErrRecordNotFound {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, &RtnJsonBase{Msg: "指定图片不存在"})
				return
			}
			log.Printf("Image replace update image record error: %v", err.Error())
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if err := ctx.SaveUploadedFile(fh, fPath); err != nil {
			log.Printf("Image replace save uploaded file error: %v", err.Error())
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		var img *models.Image
		if img, err = models.GetImageByName(name); err != nil {
			log.Printf("Image replace load new image info error: %v", err.Error())
			ctx.AbortWithStatusJSON(http.StatusBadRequest, &RtnJsonBase{Msg: "加载信息失败"})
			return
		}
		ctx.JSON(http.StatusOK, &RtnJsonImage{
			RtnJsonBase: RtnJsonBase{Msg: "ok"},
			Data:        img,
		})
		return
	}
	ctx.AbortWithStatusJSON(http.StatusBadRequest, &RtnJsonBase{Msg: "缺失参数"})
}

// 文件后缀检查
func isImageExtValid(ext string) bool {
	switch ext {
	// TODO 表驱动替换硬编码
	case ".jpg", ".png", ".bmp":
		return true
	}
	return false
}

func genImageHash(v string) string {
	rand.Seed(time.Now().Unix())
	hash := fmt.Sprintf(
		"%x",
		md5.Sum([]byte(fmt.Sprintf("%s%d", "imageserver"+v, rand.Int()))),
	)
	return hash
}

func imageValidCheck(f io.Reader, t string) (bool, error) {
	var err error
	switch t {
	case "jpg", "jpeg":
		_, err = jpeg.DecodeConfig(f)
	case "png":
		_, err = png.DecodeConfig(f)
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func imageDecode(path string) (*image.Image, error) {
	f, err := os.Open(path)

	ext := filepath.Ext(path)
	if err != nil {
		return nil, err
	}

	var img image.Image

	switch ext {
	case ".jpg", "jpeg":
		img, err = jpeg.Decode(f)
	case ".png":
		img, err = png.Decode(f)
	default:
		return nil, errors.New("unknown image type")
	}
	if err != nil {
		return nil, err
	}
	return &img, nil
}
