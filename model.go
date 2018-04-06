package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"time"
)

type User struct {
	gorm.Model
	Uid  uint32 `gorm:"unique_index;not null"`
	Name string `gorm:"size:64;not null"`
}

type Store struct {
	gorm.Model
	User uint32     `gorm:"index"`
	Size int64      `gorm:"index"`
	Age  *time.Time `gorm:"index"`
	//FQName string     `gorm:"size:4096;unique;not null"`
	FQName string `gorm:"size:4096;not null"`
}

type FileStore struct {
	Store
	// Hash64k     []byte
	// HashLast64k []byte
	// Hash64M     []byte
	// HashLast64M []byte
	// HashAll     []byte
}

type DirStore struct {
	Store
	NumFiles uint64 `gorm:"index:num_files"`
}
