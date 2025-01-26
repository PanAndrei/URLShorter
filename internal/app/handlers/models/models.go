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

	if req.Original != "" && req.Original != "null" {
		url = req.Original
	} else {
		url = req.URL
	}

	return repo.URL{
		FullURL: url,
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
	}
}

func (r *APIResponse) FromURLs(reps []repo.URL, host string) []APIResponse {
	responses := make([]APIResponse, 0, len(reps))
	for _, rep := range reps {
		responses = append(responses, r.FromURL(rep, host))
	}
	return responses
}

type ButchRequest struct {
	Short    string `json:"short_url"`
	Original string `json:"original_url"`
}

func (b *ButchRequest) FromURLs(reps []*repo.URL, host string) []ButchRequest {
	resp := make([]ButchRequest, 0, len(reps))

	for i := range reps {
		n := ButchRequest{
			Short:    host + "/" + reps[i].ShortURL,
			Original: reps[i].FullURL,
		}

		resp = append(resp, n)
	}

	return resp
}
