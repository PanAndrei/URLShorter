package repository

import (
	"errors"
)

type Repository interface {
	SaveURL(u *URL)
	LoadURL(u *URL) (r *URL, err error)
	IsUniqueShort(s string) bool
}

type Store struct {
	s map[string]string // + race
}

func NewStore() *Store {
	return &Store{
		s: make(map[string]string),
	}
}

type URL struct {
	FullURL  string
	ShortURL string
}

var (
	ErrURLNotFound = errors.New("url not found")
)

func newErrURLNotFound() error {
	return ErrURLNotFound
}

func (store *Store) SaveURL(u *URL) {
	store.s[u.FullURL] = u.ShortURL
}

func (store *Store) LoadURL(u *URL) (r *URL, err error) {
	if u.FullURL == "" && u.ShortURL == "" {
		return nil, newErrURLNotFound() // empty request
	} else if u.ShortURL == "" {
		return store.loadByFullURL(u)
	} else if u.FullURL == "" {
		return store.loadByShortURL(u)
	}

	return u, nil
}

func (store *Store) IsUniqueShort(s string) bool {
	for _, v := range store.s {
		if v == s {
			return false
		}
	}

	return true
}

func (store *Store) loadByFullURL(u *URL) (r *URL, err error) {
	v, ok := store.s[u.FullURL]

	if !ok {
		return nil, newErrURLNotFound()
	}

	u.ShortURL = v
	return u, nil
}

func (store *Store) loadByShortURL(u *URL) (r *URL, err error) {
	for k, v := range store.s {
		if u.ShortURL == v {
			u.FullURL = k
			return u, nil
		}
	}

	return nil, newErrURLNotFound()
}
