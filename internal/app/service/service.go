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
	SetShortURL(url *repo.URL) (u *repo.URL)
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

func (serv *Shorter) SetShortURL(url *repo.URL) (u *repo.URL) {
	newU, err := serv.store.LoadURL(url)

	if err != nil {
		short := serv.generateUniqAdress()
		url.ShortURL = short
		serv.store.SaveURL(url)
		return url
	}

	return newU
}

func (serv *Shorter) GetFullURL(url *repo.URL) (u *repo.URL, err error) {
	newU, err := serv.store.LoadURL(url)

	if err != nil {
		return nil, err
	}

	return newU, nil
}

func (serv *Shorter) generateUniqAdress() string {
	b := make([]byte, adressLenght)

	for {
		for i := range b {
			b[i] = charset[rand.Intn(len(charset))]
		}

		if !serv.store.IsUniqueShort(string(b)) {
			b = make([]byte, adressLenght)
		} else {
			break
		}
	}

	return string(b)
}

func (serv *Shorter) Ping() error {
	return serv.store.Ping()
}

func (serv *Shorter) BatchURLs(urls *[]repo.URL) (u *[]repo.URL, err error) {
	urs := make([]*repo.URL, len(*urls))

	for _, v := range *urls {
		v.ShortURL = serv.generateUniqAdress()
		urs = append(urs, &v)
	}

	if er := serv.store.BatchURLS(urs); er != nil {
		return nil, er
	}

	return urls, nil
}
