package services

import (
	"math/rand"
)

const (
	adressLenght = 8
	charset      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" // как бы тут через range и askii покрасивее
)

func SaveURL(url string, urls *map[string]string) (short string) {
	v, ok := (*urls)[url]

	if !ok {
		short = generateUniqAdress(adressLenght, urls)
		(*urls)[url] = short
	} else {
		short = v
	}

	return
}

func LoadURL(short string, urls *map[string]string) (url string, ok bool) {
	for k, v := range *urls {
		if short == v {
			url = k
			ok = true
			break
		}
	}

	return
}

func generateUniqAdress(length int, urls *map[string]string) string {
	b := make([]byte, length)

	for {
		for i := range b {
			b[i] = charset[rand.Intn(len(charset))]
		}

		_, ok := (*urls)[string(b)]

		if ok {
			b = make([]byte, length)
		} else {
			break
		}

	}

	return string(b)
}
