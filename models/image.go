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
	if err := db.Create(&img).Error; err != nil {
		return err
	}
	return nil
}

func removeImageInStorage(path string) error {
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
	rmImg := &Image{}
	if err := db.Where("id = ?", id).First(rmImg).Error; err != nil {
		return err
	}
	if err := removeImageInStorage(rmImg.Path); err != nil {
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
	if err := removeImageInStorage(rmImg.Path); err != nil {
		return err
	}
	if err := db.Delete(rmImg).Error; err != nil {
		return err
	}
	return nil
}
