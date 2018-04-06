package main

import (
	"errors"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/gnewton/grater"
)

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

	//recurseDirs("/home/gnewton", c, dirInfoChannel, 0)
	//recurseDirs("/home", c, dirInfoChannel, 0)
	recurseDirs("/", c, dirInfoChannel, 0)

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

var counter uint64 = 0

func handleDirInfo(c chan *DirInfo, wg *sync.WaitGroup) {
	for dirInfo := range c {
		if dirInfoPersistFilter(dirInfo) {
			persistDirInfo(dirInfo)
		}
	}
	log.Println("Done handleDirInfo")
	wg.Done()
}

func fileHandler(id int, c chan *DirFiles, wg *sync.WaitGroup) {
	//log.Println("-- START handler id=", id)
	n := 0
	var bytes int64 = 0
	for filo := range c {
		for i, _ := range filo.files {

			file := filo.files[i]
			if file.Size() == 0 {
				continue
			}
			bytes += file.Size()
			if file.IsDir() {
				continue
			}

			filename := filo.dir + "/" + file.Name()
			//log.Println("filehansker:", filename)
			fi, err := os.Lstat(filename)
			if err != nil {
				log.Println(err)
				continue
			}
			if ignore(filename, fi) {
				continue
			}
			switch mode := fi.Mode(); {
			case mode.IsRegular():
			default:
				continue
			}
			n++
			if fileInfoPersistFilter(fi) {
				//persistFileInfo(f, filename, fi)
				persistFileInfo(filename, fi)
			}
		}
	}
	//log.Println("-- handler id=", id, n, bytes)
	wg.Done()
}

func recurseDirs(dir string, c chan *DirFiles, dirInfoChannel chan *DirInfo, depth int) {

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
					recurseDirs(dir+f.Name(), c, dirInfoChannel, depth+1)
				} else {
					recurseDirs(dir+"/"+f.Name(), c, dirInfoChannel, depth+1)
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

func init() {
	uidUserMap = make(map[uint32]string)
	numCpus := runtime.NumCPU()
	if numCpus/4 > numFileHandlers {
		numFileHandlers = numCpus / 4
	}

}
