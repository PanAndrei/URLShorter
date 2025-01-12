package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	gzp "URLShorter/internal/app/compress"
	cnfg "URLShorter/internal/app/handlers/config"
	models "URLShorter/internal/app/handlers/models"
	log "URLShorter/internal/app/logger"
	repo "URLShorter/internal/app/repository"
	sht "URLShorter/internal/app/service"
)

func Serve(cnf cnfg.Config, sht sht.Short, db *repo.SQLStorage) error {
	h := NewHandlers(sht, cnf, db)
	r := chi.NewRouter()
	r.Use(log.WithLoggingRequest)
	r.Use(gzp.WithGzipCompression)
	r.Use(gzp.WithGzipDecompression)

	r.Post("/", h.mainPostHandler)
	r.Post("/api/shorten", h.apiShortenHandler)
	r.Get("/{i}", h.mainGetHandler)
	r.Get("/ping", h.pingDB)

	srv := &http.Server{
		Addr:    cnf.ServerAdress,
		Handler: r,
	}

	return srv.ListenAndServe()
}

type handlers struct {
	shorter sht.Short
	config  cnfg.Config
	db      *repo.SQLStorage // temp
}

func NewHandlers(shorter sht.Short, config cnfg.Config, db *repo.SQLStorage) *handlers {
	return &handlers{
		shorter: shorter,
		config:  config,
		db:      db,
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
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := h.db.Open(); err != nil {
		fmt.Print(err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := h.db.DB.PingContext(ctx); err != nil {
		fmt.Print(err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
	h.db.Close()
}
