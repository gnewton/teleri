package main

import (
	"crypto/sha1"
	"errors"
	"io"
	"log"
	"os"
	"strings"
	"syscall"
	"time"
)

// Makes sha1 of first n bytes
func makeSha1(f *os.File, fileSize int64, n int64) ([]byte, error) {
	if f == nil {
		return nil, errors.New("File is nil")
	}

	if fileSize == 0 || n == 0 {
		var b []byte
		return b, nil
	}
	h := sha1.New()
	if fileSize < n {
		if _, err := io.Copy(h, f); err != nil {
			log.Println(err)
			return nil, err
		}
	} else {
		if _, err := io.CopyN(h, f, n); err != nil {
			log.Println(err)
			return nil, err
		}
	}
	return h.Sum(nil), nil

}

func setUid() {
	var callerUid int = syscall.Getuid()
	uids[callerUid] = struct{}{}
}

func tooOld(t time.Time) bool {
	now := time.Now()

	ellapsed := now.Sub(t)

	if float64(ellapsed.Minutes()) > float64(ageLimit.Minutes()) {
		//log.Println(t, now, ellapsed)
		return true
	}

	return false
}

func ignore(f string, fi os.FileInfo) bool {
	for i, _ := range prefixes {
		if strings.HasPrefix(f, prefixes[i]) {
			return true
		}
	}

	return !checkUid(fi)
}

func checkUid(fi os.FileInfo) bool {
	if len(uids) == 0 {
		return false
	}
	fuid := fi.Sys().(*syscall.Stat_t).Uid

	_, contains := uids[int(fuid)]
	return contains
}
