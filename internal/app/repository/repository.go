package repository

import (
	"errors"
	"sync"
)

type Store struct {
	mux *sync.Mutex
	s   map[string]string
}

func NewStore() *Store {
	return &Store{
		mux: &sync.Mutex{},
		s:   make(map[string]string),
	}
}

type URL struct {
	FullURL  string `json:"originalUrl"`
	ShortURL string `json:"shortUrl"`
	ID       string `json:"id"`
	UUID     int    `json:"uuid"`
}

var (
	ErrURLNotFound = errors.New("url not found")
)

func newErrURLNotFound() error {
	return ErrURLNotFound
}

func (store *Store) SaveURL(u *URL) error {
	store.mux.Lock()
	defer store.mux.Unlock()

	_, err := store.loadByFullURL(u)

	if err == nil {
		return newErrURLAlreadyExists()
	}

	store.s[u.ShortURL] = u.FullURL

	return nil
}

func (store *Store) LoadURL(u *URL) (r *URL, err error) {
	store.mux.Lock()
	defer store.mux.Unlock()

	if u.FullURL == "" && u.ShortURL == "" {
		return nil, newErrURLNotFound() // empty request
	} else if u.ShortURL == "" {
		return store.loadByFullURL(u)
	} else if u.FullURL == "" {
		return store.loadByShortURL(u)
	}

	return u, nil
}

func (store *Store) loadByFullURL(u *URL) (r *URL, err error) {
	for k, v := range store.s {
		if v == u.FullURL {
			u.ShortURL = k
			return u, nil
		}
	}

	return nil, newErrURLNotFound()
}

func (store *Store) loadByShortURL(u *URL) (r *URL, err error) {
	k, ok := store.s[u.ShortURL]

	if ok {
		u.FullURL = k
		return u, nil
	}

	return nil, newErrURLNotFound()
}

func (store *Store) Ping() error {
	return nil
}

func (store *Store) BatchURLS(urls []*URL) error {
	for _, u := range urls {
		store.SaveURL(u)
	}

	return nil
}
