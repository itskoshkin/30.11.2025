package filestore

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	"github.com/spf13/viper"

	"link-availability-checker/internal/config"
	"link-availability-checker/internal/models"
	"link-availability-checker/internal/utils/closer"
)

type FileStore struct {
	path    string
	file    *os.File
	mutex   sync.RWMutex
	counter uint64 // фещьшс
}

var ErrSetNotFound = errors.New("set not found")

func NewFileStorer() (*FileStore, error) {
	fs := &FileStore{path: viper.GetString(config.LinksFilePath)}

	last, err := fs.getLastSetNumberFromFile()
	if err != nil {
		return nil, fmt.Errorf("failed to get last set number: %w", err)
	}

	file, err := os.OpenFile(fs.path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open/create links file: %w", err)
	}
	fs.file = file

	atomic.StoreUint64(&fs.counter, uint64(last))
	return fs, nil
}

func (fs *FileStore) AppendSet(set *models.Set) (int, error) {
	set.Number = int(atomic.AddUint64(&fs.counter, 1))

	bytes, err := json.Marshal(set)
	if err != nil {
		return 0, fmt.Errorf("marshal set: %w", err)
	}

	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	if _, err = fs.file.WriteString(string(bytes) + "\n"); err != nil {
		return 0, fmt.Errorf("failed to append set to links file: %w", err)
	}

	if err = fs.file.Sync(); err != nil {
		return 0, fmt.Errorf("failed to sync links file after append: %w", err)
	}

	return set.Number, nil
}

func (fs *FileStore) FindSet(number int) (*models.Set, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	file, err := os.Open(fs.path) // Opening new file descriptor to avoid interfering with appends
	if err != nil {
		return nil, fmt.Errorf("open file for find: %w", err)
	}
	defer closer.Close(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var set models.Set
		if err = json.Unmarshal(scanner.Bytes(), &set); err != nil {
			continue //TODO: Comment
		}
		if set.Number == number {
			return &set, nil
		}
	}

	if err = scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan error in find: %w", err) //TODO: Malformed 2 will break if we looked for correct 3?
	}

	return nil, ErrSetNotFound
}

func (fs *FileStore) GetLastSetNumber() (int, error) {
	return int(atomic.LoadUint64(&fs.counter)), nil
}

func (fs *FileStore) getLastSetNumberFromFile() (int, error) {
	file, err := os.Open(fs.path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil // we created new file if none existed or fell down if was unable to, but double check
		}
		return 0, err
	}
	defer closer.Close(file)

	scanner := bufio.NewScanner(file)
	last := 0
	for scanner.Scan() {
		var set models.Set
		if err = json.Unmarshal(scanner.Bytes(), &set); err != nil {
			continue //TODO: Comment
		}
		if set.Number > last {
			last = set.Number
		}
	}

	if err = scanner.Err(); err != nil {
		return 0, err
	}

	return last, nil
}
