package main

import (
	"github.com/jinzhu/gorm"
)

func dbInit(dbfilename string) (*gorm.DB, error) {
	db, err := gorm.Open("sqlite3", dbfilename)
	if err != nil {
		return nil, err
	}

	// Migrate the schema
	db.AutoMigrate(&User{})
	db.AutoMigrate(&FileStore{})
	db.AutoMigrate(&DirStore{})

	return db, nil
}
