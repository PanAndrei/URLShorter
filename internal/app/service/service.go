package services

import (
	"math/rand"

	repo "URLShorter/internal/app/repository"
)

const (
	adressLenght = 8
	charset      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type Short interface {
	SetShortURL(url *repo.URL) (u *repo.URL, err error)
	GetFullURL(url *repo.URL) (u *repo.URL, err error)
	Ping() error
	BatchURLs(urls *[]repo.URL) (u *[]repo.URL, err error)
}

type Shorter struct {
	store repo.Repository
}

func NewShorter(store repo.Repository) *Shorter {
	return &Shorter{
		store: store,
	}
}

func (serv *Shorter) SetShortURL(url *repo.URL) (u *repo.URL, err error) {
	short := serv.generateUniqAdress()
	url.ShortURL = short
	e := serv.store.SaveURL(url)

	if e != nil {
		u, err := serv.store.LoadURL(url)

		if err != nil {
			return nil, err
		}

		return u, repo.ErrURLAlreadyExists
	}

	return url, nil
}

func (serv *Shorter) GetFullURL(url *repo.URL) (u *repo.URL, err error) {
	newU, err := serv.store.LoadURL(url)
	println("gggg2", newU.FullURL, newU.ShortURL)
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

func (serv *Shorter) Ping() error {
	return serv.store.Ping()
}

func (serv *Shorter) BatchURLs(urls *[]repo.URL) (u *[]repo.URL, err error) {
	urs := make([]*repo.URL, 0, len(*urls))
	for i := range *urls {
		(*urls)[i].ShortURL = serv.generateUniqAdress()
		urs = append(urs, &(*urls)[i])
	}

	if er := serv.store.BatchURLS(urs); er != nil {
		return nil, er
	}

	return urls, nil
}
