package main

import (
	"fmt"
	"log"
	"os"
	//"syscall"
)

func persistFileInfo(f *os.File, filename string, fi os.FileInfo) error {
	//hash, err = makeMd5(f, fi.Size(), fileHeadSize)
	hash, err := makeSha1(f, fi.Size(), fileHeadSize)
	if err != nil {
		return err
	}
	hashStr := fmt.Sprintf("%x", hash)

	fuid := getFileUid(fi)

	log.Println("size=", fi.Size(),
		"modTime=", fi.ModTime(),
		"uid=", fuid,
		"user=", lookupUser(fuid),
		"path=", filename,
		"hash=", hashStr)
	return nil
}

func fileInfoPersistFilter(fi os.FileInfo) bool {
	return tooOld(fi.ModTime()) || fi.Size() > fileSizeLimit
}

func dirInfoPersistFilter(di *DirInfo) bool {
	return di.numfiles > dirNumLimit || di.size > dirSizeLimit || tooOld(di.fileInfo.ModTime())
}

func persistDirInfo(di *DirInfo) error {
	fuid := getFileUid(di.fileInfo)
	log.Println(">>>> size=",
		di.size,
		"numFiles=",
		di.numfiles,
		"user=", lookupUser(fuid),
		"path=", di.path)
	return nil
}
