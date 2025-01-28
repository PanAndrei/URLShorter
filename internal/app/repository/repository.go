package repository

import (
	"context"
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
	FullURL   string `json:"originalUrl"`
	ShortURL  string `json:"shortUrl"`
	UUID      string `json:"user_id"`
	IsDeleted bool
}

var (
	ErrURLNotFound = errors.New("url not found")
)

func newErrURLNotFound() error {
	return ErrURLNotFound
}

func (store *Store) SaveURL(_ context.Context, u *URL) (*URL, error) {
	store.mux.Lock()
	defer store.mux.Unlock()

	store.s[u.ShortURL] = u.FullURL

	return nil, nil
}

func (store *Store) LoadURL(_ context.Context, u *URL) (r *URL, err error) {
	store.mux.Lock()
	defer store.mux.Unlock()

	return store.loadByShortURL(u)
}

func (store *Store) loadByShortURL(u *URL) (r *URL, err error) {
	k, ok := store.s[u.ShortURL]

	if ok {
		u.FullURL = k
		return u, nil
	}

	return nil, newErrURLNotFound()
}

func (store *Store) Ping(_ context.Context) error {
	return nil
}

func (store *Store) BatchURLS(ctx context.Context, urls []*URL) error {
	for _, u := range urls {
		store.SaveURL(ctx, u)
	}

	return nil
}

func (store *Store) GetByUID(ctx context.Context, id string) ([]*URL, error) {
	var urls []*URL

	return urls, nil
}

func (store *Store) DeleteURLs(ctx context.Context, u []*URL) error {
	return nil
}
