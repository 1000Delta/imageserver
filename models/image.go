package models

import (
	"os"
	"time"
)

type Image struct {
	ID        int       `json:"id,omitempty"`
	Name      string    `json:"name,omitempty" gorm:"unique"`
	Ext       string    `json:"ext"`
	Path      string    `json:"path,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

func GetImageList() ([]*Image, error) {
	db, err := connDB()
	if err != nil {
		return []*Image{}, err
	}
	defer db.Close()
	var images []*Image
	if err := db.Select("id, name, path, updated_at").Find(&images).Error; err != nil {
		return []*Image{}, err
	}
	return images, nil
}

func GetImageByID(id int) (*Image, error) {
	db, err := connDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	i := &Image{ID: id}
	if err := db.Where(i).First(i).Error; err != nil {
		return nil, err
	}
	return i, nil
}

func GetImageByName(name string) (*Image, error) {
	db, err := connDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	i := &Image{Name: name}
	if err := db.Where(i).First(i).Error; err != nil {
		return nil, err
	}
	return i, nil
}

func SaveNewImage(img *Image) error {
	db, err := connDB()
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.Create(&img).Error; err != nil {
		return err
	}
	return nil
}

func RemoveImageInStorage(path string) error {
	err := os.Remove(path)
	if err != nil {
		return err
	}
	return nil
}

func RemoveImageByID(id int) error {
	db, err := connDB()
	if err != nil {
		return err
	}
	defer db.Close()
	rmImg := &Image{}
	if err := db.Where("id = ?", id).First(rmImg).Error; err != nil {
		return err
	}
	if err := RemoveImageInStorage("." + rmImg.Path); err != nil { // 替换path网络路径为本地相对路径
		return err
	}
	if err := db.Delete(rmImg).Error; err != nil {
		return err
	}
	return nil
}

func RemoveImageByName(name string) error {
	db, err := connDB()
	if err != nil {
		return err
	}
	rmImg := &Image{}
	if err := db.Where("name = ?", name).First(rmImg).Error; err != nil {
		return err
	}
	if err := RemoveImageInStorage(rmImg.Path); err != nil {
		return err
	}
	if err := db.Delete(rmImg).Error; err != nil {
		return err
	}
	return nil
}

func ReplaceImageByName(name, path string) error {
	db, err := connDB()
	if err != nil {
		return err
	}
	defer db.Close()
	img := &Image{}
	if err := db.Where("name = ?", name).First(&img).Error; err != nil {
		return err
	}
	// 删除原图片
	if err := RemoveImageInStorage("." + img.Path); err != nil {
		return err
	}
	// 更新图片路径
	img.Path = path
	if err := db.Where(&img).Update("path").Error; err != nil {
		return err
	}
	return nil
}
