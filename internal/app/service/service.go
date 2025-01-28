package services

import (
	"context"
	"math/rand"

	repo "URLShorter/internal/app/repository"
)

const (
	adressLenght = 8
	charset      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type Short interface {
	SetShortURL(ctx context.Context, url *repo.URL) (u *repo.URL, err error)
	GetFullURL(ctx context.Context, url *repo.URL) (u *repo.URL, err error)
	Ping(ctx context.Context) error
	BatchURLs(ctx context.Context, urls *[]repo.URL) (u *[]repo.URL, err error)
	GetByUID(ctx context.Context, id string) (u []*repo.URL, err error)
	DeleteURLs(ctx context.Context, u []*repo.URL) error
}

type Shorter struct {
	store repo.Repository
}

func NewShorter(store repo.Repository) *Shorter {
	return &Shorter{
		store: store,
	}
}

func (serv *Shorter) SetShortURL(ctx context.Context, url *repo.URL) (u *repo.URL, err error) {
	short := serv.generateUniqAdress()
	tmp := repo.URL{
		FullURL:  url.FullURL,
		ShortURL: short,
		UUID:     url.UUID,
	}

	_, e := serv.store.SaveURL(ctx, &tmp)

	if e != nil {
		loadedURL, err := serv.store.LoadURL(ctx, url)
		if err != nil {
			return nil, repo.ErrURLNotFound
		}
		return loadedURL, repo.ErrURLAlreadyExists
	}

	return &tmp, nil
}

func (serv *Shorter) GetFullURL(ctx context.Context, url *repo.URL) (u *repo.URL, err error) {
	newU, err := serv.store.LoadURL(ctx, url)
	if err != nil {
		return nil, err
	}

	return newU, nil
}

func (serv *Shorter) generateUniqAdress() string {
	b := make([]byte, adressLenght)

	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}

func (serv *Shorter) Ping(ctx context.Context) error {
	return serv.store.Ping(ctx)
}

func (serv *Shorter) BatchURLs(ctx context.Context, urls *[]repo.URL) (u *[]repo.URL, err error) {
	urs := make([]*repo.URL, 0, len(*urls))
	for i := range *urls {
		(*urls)[i].ShortURL = serv.generateUniqAdress()
		urs = append(urs, &(*urls)[i])
	}

	if er := serv.store.BatchURLS(ctx, urs); er != nil {
		return nil, er
	}

	return urls, nil
}

func (serv *Shorter) GetByUID(ctx context.Context, id string) (u []*repo.URL, err error) {
	return serv.store.GetByUID(ctx, id)
}

func (serv *Shorter) DeleteURLs(ctx context.Context, u []*repo.URL) error {
	return serv.store.DeleteURLs(ctx, u)
}
