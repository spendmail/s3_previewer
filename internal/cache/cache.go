package cache

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

type Config interface {
	GetCacheCapacity() int64
	GetCachePath() string
}

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

type LruCache struct {
	capacity int64
	queue    List
	items    map[string]*ListItem
	path     string
	logger   Logger
	mutex    sync.Mutex
}

type cacheItem struct {
	key   string
	value string
}

var (
	ErrFileWrite     = errors.New("unable to write file to filesystem")
	ErrFileRemove    = errors.New("unable to remove file from filesystem")
	ErrFileRead      = errors.New("unable to read file from filesystem")
	ErrItemNotExists = errors.New("cache item does not exist")
)

// New is a cache constructor: returns lruCache instance pointer.
func New(config Config, logger Logger) (*LruCache, error) {
	cache := LruCache{
		capacity: config.GetCacheCapacity(),
		path:     config.GetCachePath(),
		queue:    NewList(),
		items:    make(map[string]*ListItem, config.GetCacheCapacity()),
		logger:   logger,
	}

	err := os.MkdirAll(cache.path, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return &cache, nil
}

// Get is a LruCache getter: returns value if exists, or error, if doesnt.
func (l *LruCache) Get(key string) ([]byte, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	item, exists := l.items[key]

	// If cache element doesn't exist, return empty slice
	if !exists {
		return []byte{}, ErrItemNotExists
	}

	// If cache element exists, move it to front
	l.queue.MoveToFront(item)

	// To get actual value, interface{} needs to be casted to cacheItem
	cacheItemElement := item.Value.(cacheItem)
	filename := cacheItemElement.value

	// Reading from filesystem
	value, err := l.readFromFileSystem(filename)
	if err != nil {
		return []byte{}, fmt.Errorf("%w: %s", ErrFileRead, err)
	}

	return value, nil
}

// Set is a LruCache setter: sets or updates value, depends on whether the value exists or not.
func (l *LruCache) Set(key string, imageBytes []byte) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	listItem, exists := l.items[key]

	filename := encodeFileName(key)
	cacheItemElement := cacheItem{key, filename}

	if exists {
		// If cache element exists, move it to front
		listItem.Value = cacheItemElement
		l.queue.MoveToFront(listItem)
	} else {
		// If cache element doesn't exist, create
		listItem = l.queue.PushFront(cacheItemElement)

		// If list exceeds capacity, remove last element from list and map
		if int64(l.queue.Len()) > l.capacity {
			l.removeLastRecentUsedElement()
		}

		// Saving file to filesystem
		err := l.saveToFileSystem(filename, imageBytes)
		if err != nil {
			l.logger.Error(fmt.Errorf("%w: %s", ErrFileWrite, err))
		}
	}

	// Update map value anyway
	l.items[key] = listItem

	return nil
}

// removeLastRecentUsedElement removes LRU element from queue and file from filesystem.
func (l *LruCache) removeLastRecentUsedElement() {
	if item := l.queue.Back(); item != nil {
		backCacheItem := item.Value.(cacheItem)
		filename := backCacheItem.value

		delete(l.items, backCacheItem.key)
		l.queue.Remove(item)

		// Removing expired file from filesystem.
		err := l.removeFromFileSystem(filename)
		if err != nil {
			l.logger.Error(fmt.Errorf("%w: %s", ErrFileRemove, err))
		}
	}
}

// encodeFileName generates filename from key.
func encodeFileName(key string) string {
	return base64.StdEncoding.EncodeToString([]byte(key))
}

// saveToFileSystem writes file to filesystem.
func (l *LruCache) saveToFileSystem(filename string, bytes []byte) error {
	absFilename := filepath.Join(l.path, filename)
	return ioutil.WriteFile(absFilename, bytes, 0o600)
}

// reads file from filesystem.
func (l *LruCache) readFromFileSystem(filename string) ([]byte, error) {
	absFilename := filepath.Join(l.path, filename)
	bytes, err := ioutil.ReadFile(absFilename)
	if err != nil {
		return []byte{}, err
	}

	return bytes, nil
}

// removeFromFileSystem removes file from filesystem.
func (l *LruCache) removeFromFileSystem(filename string) error {
	absFilename := filepath.Join(l.path, filename)
	err := os.Remove(absFilename)

	return err
}

// Clear re-init lruCache instance.
func (l *LruCache) Clear() {
	l.queue = NewList()
	l.items = make(map[string]*ListItem, l.capacity)
}
