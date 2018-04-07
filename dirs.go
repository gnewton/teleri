package main

import (
	"errors"
	"io"
	"log"
	"os"
	"sync"
)

func handleDirInfo(c chan *DirInfo, wg *sync.WaitGroup) {
	for dirInfo := range c {
		if dirInfoPersistFilter(dirInfo) {
			persistDirInfo(dirInfo)
		}
	}
	log.Println("Done handleDirInfo")
	wg.Done()
}

func recurseDirs(dir string, c chan *DirFiles, dirInfoChannel chan *DirInfo, depth int, ftc FileTimeCache) {
	fileInfo, err := os.Lstat(dir)
	if err != nil {
		log.Println(err)
		return
	}
	if ignore(dir, fileInfo) {
		return
	}

	if !fileInfo.IsDir() {
		log.Println(errors.New(dir + " is not a directory"))
		return
	}

	tmp := fileInfo.ModTime()
	if !ftc.HasChanged(dir, &tmp) {
		return
	}

	//log.Println(dir)
	file, err := os.Open(dir)
	if err != nil {
		return
	}

	var files []os.FileInfo
	var numDirFiles int64 = 0
	var dirFileTotalSize int64 = 0
	for err == nil {
		files, err = file.Readdir(dirSize)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Println(err, ":::", dir)
				break
			}
		}
		chunkStrategy(c, dir, files)

		for i, _ := range files {
			counter++
			f := files[i]
			if f.IsDir() {
				if dir == "/" {
					recurseDirs(dir+f.Name(), c, dirInfoChannel, depth+1, ftc)
				} else {
					recurseDirs(dir+"/"+f.Name(), c, dirInfoChannel, depth+1, ftc)
				}
			} else {
				numDirFiles++
				dirFileTotalSize += f.Size()
			}
		}
	}

	dirInfo := DirInfo{
		path:     dir,
		numfiles: numDirFiles,
		size:     dirFileTotalSize,
		fileInfo: fileInfo,
	}

	dirInfoChannel <- &dirInfo
}

func chunkStrategy(c chan *DirFiles, dir string, files []os.FileInfo) {
	filo := new(DirFiles)
	filo.files = files
	filo.dir = dir
	c <- filo
	return
}
