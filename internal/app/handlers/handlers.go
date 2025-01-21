package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	gzp "URLShorter/internal/app/compress"
	cookies "URLShorter/internal/app/coockies"
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
	r.Use(cookies.WithCoockies)

	r.Post("/api/shorten/batch", h.batchHandler)
	r.Post("/api/shorten", h.apiShortenHandler)
	r.Post("/", h.mainPostHandler)
	r.Get("/api/user/urls", h.getButchByID)
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
	var userID string = ""
	body, err := io.ReadAll(req.Body)

	if err != nil {
		http.Error(res, "Body is empty", http.StatusBadRequest)
		return
	}

	defer req.Body.Close()

	token, ok := req.Context().Value(cookies.TokenName).(string)
	if ok {
		userID, _ = cookies.GetUID(token)
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
		FullURL: receivedURL,
		UUID:    userID,
	}

	short, err := h.shorter.SetShortURL(req.Context(), &u)

	if err != nil {
		if errors.Is(err, repo.ErrURLAlreadyExists) {
			res.WriteHeader(http.StatusConflict)
			res.Header().Set("Content-Type", "text/plain")
			res.Write([]byte(h.config.ReturnAdress + "/" + short.ShortURL))

			return
		}
		http.Error(res, "can't save url", http.StatusBadRequest)

		return
	}
	res.WriteHeader(http.StatusCreated)
	res.Header().Set("Content-Type", "text/plain")
	res.Write([]byte(h.config.ReturnAdress + "/" + short.ShortURL))
}

func (h *handlers) apiShortenHandler(res http.ResponseWriter, req *http.Request) {
	var request models.APIRequest
	var userID string = ""

	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(res, "Body is empty", http.StatusBadRequest)
		return
	}

	token, ok := req.Context().Value(cookies.TokenName).(string)
	if ok {
		userID, _ = cookies.GetUID(token)
	}

	u := request.ToURL(request)
	u.UUID = userID
	url, e := h.shorter.SetShortURL(req.Context(), &u)

	res.Header().Set("Content-Type", "application/json")

	if e != nil {
		if errors.Is(e, repo.ErrURLAlreadyExists) {

			if url != nil {
				var response models.APIResponse
				data, _ := json.Marshal(response.FromURL(*url, h.config.ReturnAdress))
				res.WriteHeader(http.StatusConflict)
				res.Write(data)
				return
			}
			http.Error(res, "can't save url", http.StatusBadRequest)
			return
		}
		http.Error(res, "Can't save URL", http.StatusBadRequest)
		return
	}
	var response models.APIResponse
	data, err := json.Marshal(response.FromURL(*url, h.config.ReturnAdress))
	if err != nil {
		http.Error(res, "Body is empty", http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusCreated)
	res.Write(data)
}

func (h *handlers) batchHandler(res http.ResponseWriter, req *http.Request) {
	var requests []models.APIRequest
	var userID string = ""

	if err := json.NewDecoder(req.Body).Decode(&requests); err != nil {
		http.Error(res, "Body is empty", http.StatusBadRequest)
		return
	}

	us := make([]repo.URL, len(requests))

	token, ok := req.Context().Value(cookies.TokenName).(string)
	if ok {
		userID, _ = cookies.GetUID(token)
	}

	for i, r := range requests {
		us[i] = repo.URL{
			FullURL: r.Original,
			UUID:    userID,
		}
	}

	urls := &us
	u, err := h.shorter.BatchURLs(req.Context(), urls)

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	var response models.APIResponse
	resp := response.FromURLs(*u, h.config.ReturnAdress)

	for i := range resp {
		resp[i].ID = requests[i].ID
	}

	data, err := json.Marshal(resp)
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
	url, err := h.shorter.GetFullURL(req.Context(), &u)

	if err != nil {
		http.Error(res, "URL not found", http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", url.FullURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *handlers) pingDB(res http.ResponseWriter, req *http.Request) {
	if err := h.shorter.Ping(req.Context()); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}

func (h *handlers) getButchByID(res http.ResponseWriter, req *http.Request) {
	var butch models.ButchRequest
	var response []models.ButchRequest

	token := req.Context().Value(cookies.TokenName).(string)
	userID, err := cookies.GetUID(token)
	if err != nil {
		userID = ""
	}

	urls, err := h.shorter.GetByUID(req.Context(), userID)

	if err != nil || len(urls) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	response = butch.FromURLs(urls, h.config.ReturnAdress)

	resp, err := json.Marshal(response)

	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	_, err = res.Write(resp)
	if err != nil {
		return
	}
}
