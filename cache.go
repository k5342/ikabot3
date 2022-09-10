package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"
)

type FileCacheBody struct {
	Updated time.Time
	Body    *AllScheduleInfo
}

type FileCache struct {
	sync.RWMutex
	WorkDir       string
	FileCacheBody *FileCacheBody
	CacheFileName string
}

func tryRestoreCache(workdir string, cacheFilename string) (*FileCacheBody, error) {
	file, err := os.Open(path.Join(workdir, fmt.Sprintf("%s.json", cacheFilename)))
	if err != nil {
		return nil, err
	}
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	var fcBody FileCacheBody
	err = json.Unmarshal(bytes, &fcBody)
	if err != nil {
		return nil, err
	}
	return &fcBody, nil
}

func NewFileCache(workdir string, cacheName string) *FileCache {
	fcBody, err := tryRestoreCache(workdir, cacheName)
	if err != nil {
		fmt.Println(err.Error())
		fcBody = &FileCacheBody{}
	}
	return &FileCache{
		WorkDir:       workdir,
		CacheFileName: cacheName,
		FileCacheBody: fcBody,
	}
}

func (fc *FileCache) Put(data *AllScheduleInfo) (persistent bool, err error) {
	fc.RWMutex.Lock()
	defer fc.RWMutex.Unlock()

	// refresh on-memory data
	fc.FileCacheBody.Updated = time.Now()
	fc.FileCacheBody.Body = data

	bytes, err := json.Marshal(*fc.FileCacheBody)
	if err != nil {
		return false, err
	}

	file, err := os.Create(filepath.Join(fc.WorkDir, fmt.Sprintf("%s.json", fc.CacheFileName)))
	if err != nil {
		return false, err
	}

	nbytes, err := file.Write(bytes)
	if len(bytes) != nbytes {
		return false, fmt.Errorf("written bytes mismatch: %d, actual: %d", len(bytes), nbytes)
	}
	return true, nil
}

func (fc *FileCache) Get() *AllScheduleInfo {
	fc.RWMutex.RLock()
	defer fc.RWMutex.RUnlock()
	return fc.FileCacheBody.Body
}

func (fc *FileCache) MaybeGet(ttl time.Duration) *AllScheduleInfo {
	fc.RWMutex.RLock()
	defer fc.RWMutex.RUnlock()
	if fc.IsExpired(ttl) {
		return nil
	} else {
		return fc.Get()
	}
}

func (fc *FileCache) IsExpired(ttl time.Duration) bool {
	fc.RWMutex.RLock()
	defer fc.RWMutex.RUnlock()
	return time.Now().After(fc.FileCacheBody.Updated.Add(ttl))
}
