package Services

import (
	"math/rand"
)

const (
	AdressLenght = 8
	charset      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" // как бы тут через range и askii покрасивее
)

func SaveURL(url string) (short string) {
	v, ok := urls[url]

	if !ok {
		short = generateUniqAdress(AdressLenght)
		urls[url] = short
	} else {
		short = v
	}

	return
}

func LoadURL(short string) (url string, ok bool) {
	for k, v := range urls {
		if short == v {
			url = k
			ok = true
		}
	}

	return
}

var urls = make(map[string]string)

func generateUniqAdress(length int) string {
	b := make([]byte, length)

	for {
		for i := range b {
			b[i] = charset[rand.Intn(len(charset))]
		}

		_, ok := urls[string(b)]

		if ok {
			b = make([]byte, length)
		} else {
			break
		}

	}

	return string(b)
}
