package main

import (
	"io"
	"log"
	"os"
	"sync"
	"time"
)

type CachedFileContent struct {
	file         string
	data         []byte
	mutext       sync.Mutex
	lastReload   int64
	lastModified int64
}

func NewCachedFile(file string) *CachedFileContent {
	v := &CachedFileContent{file, nil, sync.Mutex{}, 0, 0}
	return v
}

func (v *CachedFileContent) Get() ([]byte, bool) {
	changed := v.internalReload()
	if changed {
		log.Printf("Get %s returned %d bytes, changed => %t", v.file, len(v.data), changed)
	}
	return v.data, changed
}

func (v *CachedFileContent) FileName() string {
	return v.file
}

func (v *CachedFileContent) internalReload() bool {
	now := time.Now().Unix()
	if now-v.lastReload < 5 {
		return false
	}

	fi, err := os.Stat(v.file)
	if err != nil {
		log.Fatal("Can't stat file", v.file)
	}

	if fi.ModTime().Unix() == v.lastModified {
		// no change
		return false
	}

	v.mutext.Lock()
	defer v.mutext.Unlock()

	v.lastReload = now

	log.Printf("Reloading cached file %s, modified at %+v, length %d bytes", v.file, fi.ModTime().Local(), fi.Size())
	v.lastModified = fi.ModTime().Unix()
	v.lastReload = now
	file, err := os.Open(v.file)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	v.data = content
	log.Printf("Loaded %d bytes from cached file %s", len(v.data), v.file)
	return true
}
