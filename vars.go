package main

import (
	"os"
	"time"
)

type DirFiles struct {
	dir   string
	files []os.FileInfo
}

type DirInfo struct {
	path     string
	numfiles int64
	size     int64
	fileInfo os.FileInfo
}

var counter uint64 = 0

var numFileHandlers = 3
var dirSize = 100
var fileHeadSize int64 = 1024 * 64 * 1024

// 1GB
//var fileSizeLimit int64 = 1024 * 1024 * 1000
var fileSizeLimit int64 = 1024 * 1

// 4 months (120 days)
var ageLimit time.Duration = time.Hour * 24 * 30 * 12
var oldFileSizeLimit int64 = 1024 * 1024 * 500

// 20 thousand files
var dirNumLimit int64 = 2 * 1000

//var dirNumLimit int64 = 2000

// 5GB
var dirSizeLimit int64 = 1024 * 1024 * 10 * 1000

//var dirSizeLimit int64 = 1024 * 1024 * 10

var prefixes = []string{"/proc", "/sys", "/cdrom", "/lost+found", "/media", "/run", "/var"}

// uids limit what files and dirs get persisted, not what dirs are to be crawled...**Not implemented**
var uids map[int]struct{}
