package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
)

var (
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

func init() {
	db, err := gorm.Open("sqlite3", "./data/data.db")
	if err != nil {
		log.Fatalf("Fatal error: %v", err.Error())
	}
	defer db.Close()

	// 在此注册模型
	db.AutoMigrate(&Image{})
}

func connDB() (*gorm.DB, error) {
	db, err := gorm.Open("sqlite3", "./data/data.db")
	if err != nil {
		return nil, err
	}
	return db, nil
}
