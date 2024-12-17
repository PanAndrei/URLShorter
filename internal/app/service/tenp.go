package services

import (
	"math/rand"

	repo "URLShorter/internal/app/repository"
)

const (
	adressLenght = 8
	charset      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" // как бы тут через range и askii покрасивее
)

type Short interface {
	SetShortURL(url *repo.URL) (u *repo.URL)
	GetFullURL(url *repo.URL) (u *repo.URL, err error)
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

	// for {
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	// url := repo.URL{ShortURL: string(b)}
	// _, err := serv.store.LoadURL(&url)

	// if err == nil {
	// 	b = make([]byte, adressLenght)
	// } else {
	// 	break
	// }
	// }

	return string(b)
}




package repository

import (
	"encoding/json"
	"errors"
	"os"
)

type Repository interface {
	SaveURL(u *URL)
	LoadURL(u *URL) (r *URL, err error)
	// IsUniqueShort(s string) bool
	Close()
}

type Store struct {
	file    *os.File
	encoder *json.Encoder
	decoder *json.Decoder
}

func NewStore(fileName string) (*Store, error) {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0666)

	if err != nil {
		return nil, err
	}

	return &Store{
		file:    file,
		encoder: json.NewEncoder(file),
		decoder: json.NewDecoder(file),
	}, nil
}

type URLData struct {
	Data []URL `json:"data"`
}

type URL struct {
	FullURL  string `json:"originalUrl"`
	ShortURL string `json:"shortUrl"`
	UUID     int    `json:"uuid"`
}

var (
	ErrURLNotFound = errors.New("url not found")
)

func newErrURLNotFound() error {
	return ErrURLNotFound
}

func (store *Store) SaveURL(u *URL) {
	if _, err := store.file.Seek(0, 0); err != nil {
		newErrURLNotFound()
	}

	data := &URLData{Data: make([]URL, 0)}
	if err := store.decoder.Decode(data); err != nil {
		newErrURLNotFound()
	}

	uuid := len(data.Data) + 1
	u.UUID = uuid
	data.Data = append(data.Data, *u)

	if _, err := store.file.Seek(0, 0); err != nil {
		newErrURLNotFound()
	}
	store.encoder.Encode(data)
}

func (store *Store) LoadURL(u *URL) (r *URL, err error) {
	data := &URLData{Data: make([]URL, 0)}
	
	if err := store.decoder.Decode(data); err != nil {
		return nil, newErrURLNotFound()
	}

	for _, ur := range data.Data {
		if ur.ShortURL == u.ShortURL {
			r = &ur
			return r, nil
		}

		if ur.FullURL == u.FullURL {
			r = &ur
			return r, nil
		}
	}

	return nil, newErrURLNotFound()
}

func (store *Store) Close() {
	store.file.Close()
}