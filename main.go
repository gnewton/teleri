package main

import (
	"errors"
	"io"
	"log"
	"os"
	"sync"
	"syscall"
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

var num = 5
var dirSize = 40
var fileHeadSize int64 = 1024 * 64 * 1024

// 1GB
var fileSizeLimit int64 = 1024 * 1024 * 1000
var oldFileSizeLimit int64 = 1024 * 1024 * 500

// 5GB
var dirSizeLimit int64 = 1024 * 1024 * 10 * 1000

// 4 months (120 days)
var ageLimit time.Duration = time.Hour * 24 * 30 * 4

// 20 thousand files
var dirNumLimit int64 = 20 * 1000

var prefixes = []string{"/proc", "/sys", "/cdrom", "/lost+found", "/media", "/run", "/var"}

var uids map[int]struct{}

func main() {
	uids = make(map[int]struct{}, 1)
	setUid()

	var wg sync.WaitGroup
	c := make(chan *DirFiles, 100)
	dirInfoChannel := make(chan *DirInfo, 100)

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	for i := 0; i < num; i++ {
		wg.Add(1)
		go handler(i, c, &wg)
	}
	go handleDirInfo(dirInfoChannel, &wg)
	wg.Add(1)

	//recurseDirs("/home/newtong", c, dirInfoChannel, 0)
	//recurseDirs("/home", c, dirInfoChannel, 0)
	recurseDirs("/", c, dirInfoChannel, 0)
	log.Println("Closing c")
	close(c)
	log.Println("Closing dirInfoChannel")
	close(dirInfoChannel)

	log.Println("Waiting")
	wg.Wait()
	log.Println("Done waiting")
	log.Println(counter)
}

var counter int64 = 0

func handleDirInfo(c chan *DirInfo, wg *sync.WaitGroup) {
	for dirInfo := range c {
		if dirInfo.numfiles > dirNumLimit || dirInfo.size > dirSizeLimit || tooOld(dirInfo.fileInfo.ModTime()) {
			// persist
			//log.Println(dirInfo)
		}
	}
	log.Println("Done handleDirInfo")
	wg.Done()
}

// func shaFile(file string)[]bytes{
// 	f, err := os.Open(file)
// 	if err != nil {
// 		log.Println(err)
// 		f.Close()
// 		return nil
// 	}

// 	h := sha1.New()
// 	if _, err := io.Copy(h, f); err != nil {
// 		log.Println(err)
// 	}

// 	//fmt.Printf("% x", h.Sum(nil))
// 	h.Sum(nil)
// 	err = f.Close()
// 	if err != nil {
// 		log.Println(err)
// 	}
// }

func handler(id int, c chan *DirFiles, wg *sync.WaitGroup) {
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

			fi, err := os.Lstat(filename)
			if ignore(filename, fi) {
				continue
			}
			switch mode := fi.Mode(); {
			case mode.IsRegular():
			default:
				continue
			}

			f, err := os.Open(filename)
			if err != nil {
				//log.Println(err)
				f.Close()
				continue
			}

			var hash []byte
			hash, err = makeSha1(f, fi.Size(), fileHeadSize)

			if err != nil {
				err = f.Close()
				continue
			}

			err = f.Close()
			if err != nil {
				log.Println(err)
			}
			n++
			handleFile(filename, fi, hash)
		}
	}
	//log.Println("-- handler id=", id, n, bytes)
	wg.Done()
}

func handleFile(filename string, fi os.FileInfo, hash []byte) {
	_ = fi.Sys().(*syscall.Stat_t)
	//log.Println("++++++++++++++++++++++++", int(stat.Uid))

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
	if numDirFiles > 1000 || dirFileTotalSize > 100000000 {
		//log.Println(dir, numDirFiles, dirFileTotalSize/1000/1000, "M")
	}

	dirInfo := DirInfo{
		path:     dir,
		numfiles: numDirFiles,
		size:     dirFileTotalSize,
		fileInfo: fileInfo,
	}

	dirInfoChannel <- &dirInfo
}

func spaces(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += " "
	}
	return s
}

func chunkStrategy(c chan *DirFiles, dir string, files []os.FileInfo) {
	filo := new(DirFiles)
	filo.files = files
	filo.dir = dir
	c <- filo
	return
}
