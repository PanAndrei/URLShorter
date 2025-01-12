package repository

import (
	hadlCnfg "URLShorter/internal/app/handlers/config"
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
		return db, nil
	}

	if config.FileStorageAdress != "" {
		return NewFileStore(config.FileStorageAdress)
	}

	return NewStore(), nil
}
