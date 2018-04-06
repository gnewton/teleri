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
	uids = make(map[int]struct{}, 1)
	setUid()

	grater, err := grater.NewGrater(time.Second, &counter)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup

	c := make(chan *DirFiles, 100)
	dirInfoChannel := make(chan *DirInfo, 100)

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	for i := 0; i < numFileHandlers; i++ {
		wg.Add(1)
		go fileHandler(i, c, &wg)
	}
	go handleDirInfo(dirInfoChannel, &wg)
	wg.Add(1)

	go handleDirInfo(dirInfoChannel, &wg)
	wg.Add(1)

	dirs := []string{"/home", "/usr"}

	for i := 0; i < len(dirs)-1; i++ {
		log.Println(i, dirs[i])
		go recurseDirs(dirs[i], c, dirInfoChannel, 0)
	}
	recurseDirs(dirs[len(dirs)-1], c, dirInfoChannel, 0)

	//recurseDirs("/skyemci01-hpc/home/interpro-lookup-svc/data/match_db", c, dirInfoChannel, 0)

	log.Println("Closing c")
	close(c)
	log.Println("Closing dirInfoChannel")
	close(dirInfoChannel)

	log.Println("Waiting")
	wg.Wait()
	log.Println("Done waiting")
	log.Println(counter)
	grater.Stop()
}
