package repository

import (
	hadlCnfg "URLShorter/internal/app/handlers/config"
	"fmt"
)

type Repository interface {
	SaveURL(u *URL)
	LoadURL(u *URL) (r *URL, err error)
	IsUniqueShort(s string) bool
}

type StorageRouter struct{}

func NewStorageRouter() *StorageRouter {
	return &StorageRouter{}
}

func (r *StorageRouter) GetStorage(config hadlCnfg.Config) (Repository, error) {
	if config.PostgreSQLAdress != "" {
		db := NewDB(config.PostgreSQLAdress)
		if err := db.Open(); err != nil {
			return nil, fmt.Errorf("opening db: %w", err)
		}
		return db, nil

	}

	if config.FileStorageAdress != "" {
		return NewFileStore(config.FileStorageAdress)
	}

	return NewStore(), nil
}
