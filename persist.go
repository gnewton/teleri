package main

import (
	"github.com/jinzhu/gorm"
	"log"
	"os"
)

func persistFileInfo(fp *FilePersister, filename string, fi os.FileInfo) error {
	//hash, err = makeMd5(f, fi.Size(), fileHeadSize)
	//hash, err := makeSha1(f, fi.Size(), fileHeadSize)
	// if err != nil {
	// 	return err
	// }
	//hashStr := fmt.Sprintf("%x", hash)

	fuid := getFileUid(fi)

	// log.Println("size=", fi.Size(),
	// 	"modTime=", fi.ModTime(),
	// 	"uid=", fuid,
	// 	"user=", lookupUser(fuid),
	// 	"path=", filename)
	//"hash=", hashStr)

	fileStore := new(FileStore)
	fileStore.User = fuid
	fileStore.Size = fi.Size()
	fileStore.FQName = filename
	t := fi.ModTime()
	fileStore.Age = &t

	fp.save(fileStore)

	return nil
}

type FilePersister struct {
	db          *gorm.DB
	fileChannel chan []*FileStore
	n           int
	files       []*FileStore
	counter     int64
}

var max = 1000

func (fp *FilePersister) init() {
	//fp.fileChannel = make(chan []*FileStore, 20)
	fp.fileChannel = make(chan []*FileStore)

	go fileWriter(fp)
}

func (fp *FilePersister) save(fs *FileStore) {
	if fp.n == max {
		fp.fileChannel <- fp.files
		fp.n = 0
	}
	if fp.n == 0 {
		fp.files = make([]*FileStore, max)
	}

	fp.files[fp.n] = fs
	fp.n++

}

func fileWriter(fp *FilePersister) {
	for files := range fp.fileChannel {

		log.Println("fileWriter")
		tx := fp.db.Begin()

		for i, _ := range files {
			f := files[i]
			if f == nil {
				break
			}
			fp.counter++
			//log.Println("======= ", i, files[i].FQName)
			tx.Create(files[i])
		}

		err := tx.Commit()
		if err != nil {
			log.Println(err)
		}
	}
}

func (fp *FilePersister) close() {
	if fp.n > 0 {
		fp.fileChannel <- fp.files
	}
	close(fp.fileChannel)
	log.Println("counter", fp.counter)
}

func logErrors(errs []error) {
	for i := 0; i < len(errs); i++ {
		log.Println(errs[i])
	}
}

func fileInfoPersistFilter(fi os.FileInfo) bool {
	return tooOld(fi.ModTime()) && fi.Size() > oldFileSizeLimit || fi.Size() > fileSizeLimit
}

func dirInfoPersistFilter(di *DirInfo) bool {
	return di.numfiles > dirNumLimit || di.size > dirSizeLimit
}

func persistDirInfo(di *DirInfo) error {
	fuid := getFileUid(di.fileInfo)
	log.Println(">>>> size=",
		di.size,
		"numFiles=",
		di.numfiles,
		"user=", lookupUser(fuid),
		"modTime=", di.fileInfo.ModTime(),
		"path=", di.path)
	return nil
}
