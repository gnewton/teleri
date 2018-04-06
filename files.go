package main

import (
	"log"
	"os"
	"sync"
)

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
