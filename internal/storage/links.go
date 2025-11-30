package storage

import (
	"link-availability-checker/internal/models"
	"link-availability-checker/pkg/filestore"
)

type LinkStorage interface {
	SaveLinkSet(set *models.Set) (int, error)
	GetLinkSet(number int) (*models.Set, error)
}

type LinkStorageImpl struct{ fs *filestore.FileStore }

func NewLinkStorage(fs *filestore.FileStore) LinkStorage { return &LinkStorageImpl{fs: fs} }

func (s *LinkStorageImpl) SaveLinkSet(set *models.Set) (int, error) {
	return s.fs.AppendSet(set)
}

func (s *LinkStorageImpl) GetLinkSet(number int) (*models.Set, error) {
	return s.fs.FindSet(number)
}
