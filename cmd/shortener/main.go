package main

import (
	"io"
	"net/http"
	"strings"

	"github.com/PanAndrei/URLShorter/internal/app/services"
)

const (
	LocalHost = "http://localhost:8080/"
)

var urls map[string]string

func main() {
	urls = make(map[string]string)

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, mainRouteHandler)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}

func mainRouteHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		saveHandler(res, req)
	case http.MethodGet:
		loadHandler(res, req)
	default:
		http.Error(res, "Method not allowed!", http.StatusBadRequest)
	}
}

func saveHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)

	if err != nil {
		http.Error(res, "Body is empty", http.StatusBadRequest)
		return
	}

	receivedURL := strings.TrimSpace(string(body))

	lines := strings.Split(receivedURL, "\n")
	if len(lines) > 0 {
		receivedURL = strings.TrimSpace(lines[len(lines)-1])
	} else {
		http.Error(res, "Пустой боди", http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusCreated)
	res.Header().Set("Content-Type", "text/plain")
	res.Write([]byte(LocalHost + services.SaveURL(receivedURL, &urls)))
}

func loadHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}

	path := req.URL.Path
	shortURL := strings.TrimPrefix(path, "/")

	url, ok := services.LoadURL(shortURL, &urls)

	if !ok {
		http.Error(res, "URL not found", http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", url)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
