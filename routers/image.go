package routers

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/1000Delta/imageserver/models"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
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

// 图片列表返回数据
type RtnJsonImageList struct {
	RtnJsonBase
	Total int             `json:"total"`
	Data  []*models.Image `json:"data"`
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
			if err == gorm.ErrRecordNotFound {
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
			if err == gorm.ErrRecordNotFound {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}
			log.Printf("Query image by name error: %v", err.Error())
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, &RtnJsonBase{"查询图片信息失败"})
			return
		}
		ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/%s", img.Path))
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
	switch ext {
	case ".jpg", ".png", ".bmp":
	default:
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &RtnJsonBase{Msg: "文件后缀类型不支持"})
		return
	}
	// TODO 文件类型检查

	// TODO 修改路径为配置引入
	rand.Seed(time.Now().Unix())
	hash := fmt.Sprintf(
		"%x",
		md5.Sum([]byte(fmt.Sprintf("%s%d", "imageserver"+fh.Filename, rand.Int()))),
	)
	fPath := "./uploads/tmp/" + hash + ext
	err = ctx.SaveUploadedFile(fh, fPath)
	if err != nil {
		log.Printf("Save file error: %v\n", err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &RtnJsonBase{Msg: "服务器保存图片失败"})
		return
	}

	// 保存图片信息到数据库
	err = models.SaveNewImage(&models.Image{
		Name: fh.Filename,
		Ext:  ext,
		Path: fPath[1:], // 存储为绝对路径
	})

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

	ctx.JSON(http.StatusOK, &RtnJsonBase{Msg: "ok"})
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
		ctx.Status(http.StatusOK)
		return
	}
	if name, exist := ctx.GetQuery("name"); exist {
		if err := models.RemoveImageByName(name); err != nil {
			log.Printf("Remove image by name error: %v", err.Error())
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, &RtnJsonBase{Msg: "删除失败"})
			return
		}
		ctx.Status(http.StatusOK)
		return
	}
	ctx.AbortWithStatusJSON(http.StatusBadRequest, &RtnJsonBase{Msg: "缺失参数"})
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
		return nil, errors.New("Unknown image type")
	}
	if err != nil {
		return nil, err
	}
	return &img, nil
}
