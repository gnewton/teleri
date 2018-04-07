package main

import (
	"log"
	"os"
	"sync"
)

func fileHandler(fp *FilePersister, id int, c chan *DirFiles, wg *sync.WaitGroup, ftc FileTimeCache) {
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
			tmp := fi.ModTime()
			if !ftc.HasChanged(filename, &tmp) {
				return
			}

			n++
			if fileInfoPersistFilter(fi) {
				//persistFileInfo(f, filename, fi)
				persistFileInfo(fp, filename, fi)
			}
		}
	}
	//log.Println("-- handler id=", id, n, bytes)
	wg.Done()
}
