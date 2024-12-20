package models

import (
	repo "URLShorter/internal/app/repository"
)

type APIRequest struct {
	URL string `json:"url"`
}

func (r *APIRequest) ToURL(req APIRequest) repo.URL {
	return repo.URL{
		FullURL: req.URL,
	}
}

type APIResponse struct {
	Result string `json:"result"`
}

func (r *APIResponse) FromURL(rep repo.URL, host string) APIResponse {
	return APIResponse{
		Result: host + "/" + rep.ShortURL,
	}
}
