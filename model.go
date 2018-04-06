package main

import (
	"time"
)

type User struct {
	Id   uint64
	Uid  uint32
	Name string
}

type Store struct {
	Id        uint64
	ModTime   time.Time
	CheckTime time.Time
	User      uint64
	Size      uint64
	FQName    string
}

type FileStore struct {
	Store
	Hash64k     []byte
	HashLast64k []byte
	Hash64M     []byte
	HashLast64M []byte
	HashAll     []byte
}

type DirStore struct {
	Store
	NumFiles uint64
}
