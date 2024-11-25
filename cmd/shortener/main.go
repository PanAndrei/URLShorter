package main

import (
	"io"
	"net/http"
	"strings"

	"github.com/PanAndrei/URLShorter/internal/app/Services"
)

const (
	LocalHost = "http://localhost:8080/"
)

func mainHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)

	if err != nil {
		http.Error(res, "Body is empty", http.StatusBadRequest)
		return
	}

	receivedURL := string(body)

	res.WriteHeader(http.StatusCreated)
	res.Header().Set("Content-Type", "text/plain")
	res.Write([]byte(LocalHost + Services.SaveURL(receivedURL)))
}

func answerHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}

	path := req.URL.Path
	shortURL := strings.TrimPrefix(path, "/")

	url, ok := Services.LoadURL(shortURL)

	if !ok {
		http.Error(res, "URL not found", http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusTemporaryRedirect)
	res.Header().Set("Content-Type", "text/plain")
	res.Write([]byte(url))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, mainHandler)
	mux.HandleFunc(`/{id}`, answerHandler)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
