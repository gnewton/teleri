package main

import (
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

type DirFiles struct {
	dir   string
	files []os.FileInfo
}

type DirInfo struct {
	path     string
	numfiles int64
	size     int64
}

const num = 5
const dirSize = 40
const fileHeadSize = 1024 * 64 * 1024

func main() {
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
	for _ = range c {

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
			if badFilePrefix(filename) {
				continue
			}
			fileInfo, err := os.Lstat(filename)
			switch mode := fileInfo.Mode(); {
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

			_, err = makeSha1(f, fileInfo.Size(), fileHeadSize)

			if err != nil {
				err = f.Close()
				continue
			}
			//fmt.Printf("\n% x", h.Sum(nil))

			err = f.Close()
			if err != nil {
				log.Println(err)
			}
			n++
		}
	}
	//log.Println("-- handler id=", id, n, bytes)
	wg.Done()
}

func recurseDirs(dir string, c chan *DirFiles, dirInfoChannel chan *DirInfo, depth int) {
	if badFilePrefix(dir) {
		return
	}
	file, err := os.Open(dir)
	if err != nil {
		//log.Println(err)
		return
	}

	//log.Println(file)

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
		//log.Println("Chan size=", len(c))
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
	if true {
		filo := new(DirFiles)
		filo.files = files
		filo.dir = dir
		c <- filo
		return
	}

	// sort.Sort(BySize(files))

	// smallFiles := make([]os.FileInfo, 0)
	// for i, _ := range files {
	// 	//log.Println("Chan size=", len(c))
	// 	f := files[i]
	// 	if f.Size() > 1000000000 {
	// 		filo := DirFiles{
	// 			dir:   dir,
	// 			files: []os.FileInfo{f},
	// 		}
	// 		c <- &filo
	// 		log.Println("$$$$$$$$$$$$$$$$$$$$$$$", f.Name(), f.Size()/1000/1000)
	// 	} else {
	// 		smallFiles = append(smallFiles, f)
	// 	}
	// }
	// if len(smallFiles) > 0 {
	// 	filo := new(DirFiles)
	// 	filo.files = smallFiles
	// 	filo.dir = dir
	// 	c <- filo
	// }
}

func badFilePrefix(f string) bool {
	//var prefixes = []string{"/proc", "/dev", "/sys/devices", "/boot", "/cdrom", "/lost+found", "/media", "/run", "/var", "/bin"}
	var prefixes = []string{"/proc/", "/sys/", "/cdrom/", "/lost+found/", "/media/", "/run/", "/var/"}
	for i, _ := range prefixes {
		if strings.HasPrefix(f, prefixes[i]) {
			return true
		}
	}
	return false
}
