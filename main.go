package main

import (
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/gnewton/grater"
)

func init() {
	uidUserMap = make(map[uint32]string)
	numCpus := runtime.NumCPU()
	if numCpus/4 > numFileHandlers {
		numFileHandlers = numCpus / 4
	}

}

func main() {
	db, err := dbInit("foo.sql")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	uids = make(map[int]struct{}, 1)
	setUid()

	grater, err := grater.NewGrater(time.Second, &counter)
	if err != nil {
		log.Fatal(err)
	}
	defer grater.Stop()

	var wg sync.WaitGroup

	dirFilesChannel := make(chan *DirFiles, 100)
	dirInfoChannel := make(chan *DirInfo, 100)

	filePersister := new(FilePersister)
	filePersister.db = db
	filePersister.init()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	for i := 0; i < numFileHandlers; i++ {
		wg.Add(1)
		go fileHandler(filePersister, i, dirFilesChannel, &wg)
	}
	go handleDirInfo(dirInfoChannel, &wg)
	wg.Add(1)

	go handleDirInfo(dirInfoChannel, &wg)
	wg.Add(1)

	//dirs := []string{"/home", "/usr", }
	dirs := []string{"/"}

	for i := 0; i < len(dirs)-1; i++ {
		log.Println(i, dirs[i])
		go recurseDirs(dirs[i], dirFilesChannel, dirInfoChannel, 0)
	}
	recurseDirs(dirs[len(dirs)-1], dirFilesChannel, dirInfoChannel, 0)

	//recurseDirs("/skyemci01-hpc/home/interpro-lookup-svc/data/match_db", c, dirInfoChannel, 0)

	log.Println("Closing c")
	close(dirFilesChannel)
	log.Println("Closing dirInfoChannel")
	close(dirInfoChannel)

	log.Println("Waiting")
	wg.Wait()
	log.Println("Done waiting")
	log.Println(counter)

	filePersister.close()
}
