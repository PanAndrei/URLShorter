package repository

import (
	hadlCnfg "URLShorter/internal/app/handlers/config"
	"context"
)

type Repository interface {
	SaveURL(ctx context.Context, u *URL) (*URL, error)
	LoadURL(ctx context.Context, u *URL) (r *URL, err error)
	Ping(ctx context.Context) error
	BatchURLS(ctx context.Context, urls []*URL) error
	GetByUID(ctx context.Context, id string) ([]*URL, error)
	DeleteURLs(ctx context.Context, u []*URL) error
}

type StorageRouter struct{}

func NewStorageRouter() *StorageRouter {
	return &StorageRouter{}
}

func (r *StorageRouter) GetStorage(config hadlCnfg.Config) (Repository, error) {
	if config.PostgreSQLAdress != "" {
		db, err := NewDB(config.PostgreSQLAdress)
		if err == nil {
			return db, nil
		}
	}

	if config.FileStorageAdress != "" {
		return NewFileStore(config.FileStorageAdress)
	}

	return NewStore(), nil
}
