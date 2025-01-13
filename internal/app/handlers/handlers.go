package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	gzp "URLShorter/internal/app/compress"
	cnfg "URLShorter/internal/app/handlers/config"
	models "URLShorter/internal/app/handlers/models"
	log "URLShorter/internal/app/logger"
	repo "URLShorter/internal/app/repository"
	sht "URLShorter/internal/app/service"
)

func Serve(cnf cnfg.Config, sht sht.Short) error {
	h := NewHandlers(sht, cnf)
	r := chi.NewRouter()
	r.Use(log.WithLoggingRequest)
	r.Use(gzp.WithGzipCompression)
	r.Use(gzp.WithGzipDecompression)

	r.Post("/api/shorten/batch", h.batchHandler)
	r.Post("/api/shorten", h.apiShortenHandler)
	r.Post("/", h.mainPostHandler)
	r.Get("/ping", h.pingDB)
	r.Get("/{i}", h.mainGetHandler)

	srv := &http.Server{
		Addr:    cnf.ServerAdress,
		Handler: r,
	}

	return srv.ListenAndServe()
}

type handlers struct {
	shorter sht.Short
	config  cnfg.Config
}

func NewHandlers(shorter sht.Short, config cnfg.Config) *handlers {
	return &handlers{
		shorter: shorter,
		config:  config,
	}
}

func (h *handlers) mainPostHandler(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)

	if err != nil {
		http.Error(res, "Body is empty", http.StatusBadRequest)
		return
	}

	defer req.Body.Close()

	receivedURL := strings.TrimSpace(string(body))
	lines := strings.Split(receivedURL, "\n")

	if len(lines) > 0 {
		receivedURL = strings.TrimSpace(lines[len(lines)-1])
	} else {
		http.Error(res, "Body is empty", http.StatusBadRequest)
		return
	}

	u := repo.URL{
		FullURL: receivedURL,
	}

	short := h.shorter.SetShortURL(&u).ShortURL

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(h.config.ReturnAdress + "/" + short))
}

func (h *handlers) apiShortenHandler(res http.ResponseWriter, req *http.Request) {
	var request models.APIRequest

	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(res, "Body is empty", http.StatusBadRequest)
		return
	}

	u := request.ToURL(request)
	h.shorter.SetShortURL(&u)

	var response models.APIResponse
	data, err := json.Marshal(response.FromURL(u, h.config.ReturnAdress))

	if err != nil {
		http.Error(res, "Body is empty", http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	res.Write(data)
}

func (h *handlers) batchHandler(res http.ResponseWriter, req *http.Request) {
	var requests []models.APIRequest

	if err := json.NewDecoder(req.Body).Decode(&requests); err != nil {
		http.Error(res, "Body is empty", http.StatusBadRequest)
		return
	}

	us := make([]repo.URL, len(requests))

	for i, r := range requests {
		println("pr", r.URL, r.ID)

		us[i] = repo.URL{
			FullURL: r.URL,
			ID:      r.ID,
		}
	}

	urls := &us
	_, err := h.shorter.BatchURLs(urls)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	var response models.APIResponse
	data, err := json.Marshal(response.FromURLs(us, h.config.ReturnAdress))
	if err != nil {
		http.Error(res, "Marshal error", http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	res.Write(data)
}

func (h *handlers) mainGetHandler(res http.ResponseWriter, req *http.Request) {
	iStr := chi.URLParam(req, "i")

	u := repo.URL{
		ShortURL: iStr,
	}

	url, err := h.shorter.GetFullURL(&u)

	if err != nil {
		http.Error(res, "URL not found", http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", url.FullURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *handlers) pingDB(res http.ResponseWriter, req *http.Request) { // тесты
	if err := h.shorter.Ping(); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}
