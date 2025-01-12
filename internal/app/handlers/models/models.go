package models

import (
	repo "URLShorter/internal/app/repository"
)

type APIRequest struct {
	URL      string `json:"url"`          //`json:"original_url"`
	Original string `json:"original_url"` // !!
	ID       string `json:"correlation_id"`
}

func (r *APIRequest) ToURL(req APIRequest) repo.URL {
	url := ""

	if req.Original != "" {
		url = req.Original
	} else {
		url = req.URL
	}

	return repo.URL{
		FullURL: url,
		ID:      req.ID,
	}
}

func (r *APIRequest) ToURLs(reqs []APIRequest) []repo.URL {
	urls := make([]repo.URL, 0, len(reqs))
	for _, req := range reqs {
		urls = append(urls, r.ToURL(req))
	}
	return urls
}

type APIResponse struct {
	Result string `json:"result"` //`json:"short_url"`
	Short  string `json:"short_url"`
	ID     string `json:"correlation_id"`
}

func (r *APIResponse) FromURL(rep repo.URL, host string) APIResponse {
	return APIResponse{
		Result: host + "/" + rep.ShortURL,
		Short:  host + "/" + rep.ShortURL,
		ID:     rep.ID,
	}
}

func (r *APIResponse) FromURLs(reps []repo.URL, host string) []APIResponse {
	responses := make([]APIResponse, 0, len(reps))
	for _, rep := range reps {
		responses = append(responses, r.FromURL(rep, host))
	}
	return responses
}
