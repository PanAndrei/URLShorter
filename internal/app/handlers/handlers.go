package handlers

import (
	"io"
	"net/http"
	"strings"

	cnf "URLShorter/internal/app/config"
	cnfg "URLShorter/internal/app/handlers/config"
	repo "URLShorter/internal/app/repository"
	sht "URLShorter/internal/app/service"
)

func Serve(cnf cnfg.Config, sht sht.Short) error {
	h := NewHandlers(sht)
	router := newRouter(h)

	srv := &http.Server{
		Addr:    cnf.ServerAdress,
		Handler: router,
	}

	return srv.ListenAndServe()
}

type handlers struct {
	shorter sht.Short
}

func NewHandlers(shorter sht.Short) *handlers {
	return &handlers{
		shorter: shorter,
	}
}

func newRouter(h *handlers) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /", h.mainPostHandler)
	mux.HandleFunc("GET /{i}", h.mainGetHandler)

	return mux
}

func (h *handlers) mainPostHandler(res http.ResponseWriter, req *http.Request) {
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
		http.Error(res, "Body is empty", http.StatusBadRequest)
		return
	}

	u := repo.URL{
		FullUrl: receivedURL,
	}

	short := h.shorter.SetShortURL(&u).ShortURL

	res.WriteHeader(http.StatusCreated)
	res.Header().Set("Content-Type", "text/plain")
	res.Write([]byte(cnf.LocalHost + short))
}

func (h *handlers) mainGetHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}

	iStr := req.PathValue("i")
	u := repo.URL{
		ShortURL: iStr,
	}

	url, err := h.shorter.GetFullURL(&u)

	if err != nil {
		http.Error(res, "URL not found", http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", url.FullUrl)
	res.WriteHeader(http.StatusTemporaryRedirect)
}