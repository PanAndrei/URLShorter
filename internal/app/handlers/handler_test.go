package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	cnfg "URLShorter/internal/app/handlers/config"
	models "URLShorter/internal/app/handlers/models"
	repo "URLShorter/internal/app/repository"

	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockShortener struct{}

func (m *MockShortener) SetShortURL(u *repo.URL) *repo.URL {
	u.ShortURL = "m.ShortURL"

	return u
}

func (m *MockShortener) GetFullURL(u *repo.URL) (*repo.URL, error) {

	return &repo.URL{FullURL: "m.FullURL"}, nil
}

func TestMainPostHandler(t *testing.T) {
	h := NewHandlers(&MockShortener{}, cnfg.Config{}, repo.NewDB("d"))

	type set struct {
		method      string
		path        string
		contentType string
	}

	type want struct {
		responseCode int
		request      string
		contentType  string
	}

	tests := []struct {
		name string
		set  set
		want want
	}{
		{
			name: "test #1 right response",
			set: set{
				method:      http.MethodPost,
				path:        "/",
				contentType: "text/plain",
			},
			want: want{
				responseCode: http.StatusCreated,
				request:      "",
				contentType:  "text/plain",
			},
		},
		{
			name: "test #2 wrong method",
			set: set{
				method:      http.MethodGet,
				path:        "/",
				contentType: "text/plain",
			},
			want: want{
				responseCode: http.StatusBadRequest,
				request:      "",
				contentType:  "text/plain",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.set.method, test.set.path, nil)
			w := httptest.NewRecorder()
			h.mainPostHandler(w, req)
			res := w.Result()

			defer res.Body.Close()

			_, err := io.ReadAll(res.Body)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			require.NoError(t, err)
		})
	}
}

func TestMainGetHandler(t *testing.T) {
	h := NewHandlers(&MockShortener{}, cnfg.Config{}, repo.NewDB("d"))

	type set struct {
		method      string
		path        string
		contentType string
	}

	type want struct {
		responseCode int
		request      string
		contentType  string
	}

	tests := []struct {
		name string
		set  set
		want want
	}{
		{
			name: "test #1 right response",
			set: set{
				method:      http.MethodGet,
				path:        "/test",
				contentType: "text/plain",
			},
			want: want{
				responseCode: http.StatusTemporaryRedirect,
				request:      "",
				contentType:  "",
			},
		},
		{
			name: "test #2 wrong method",
			set: set{
				method:      http.MethodPost,
				path:        "/test",
				contentType: "text/plain",
			},
			want: want{
				responseCode: http.StatusBadRequest,
				request:      "",
				contentType:  "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.set.method, test.set.path, nil)
			w := httptest.NewRecorder()
			h.mainGetHandler(w, req)
			res := w.Result()

			defer res.Body.Close()

			_, err := io.ReadAll(res.Body)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			require.NoError(t, err)
		})
	}
}

func TestApishortenHandler(t *testing.T) {
	h := NewHandlers(&MockShortener{}, cnfg.Config{}, repo.NewDB("d"))

	type set struct {
		method      string
		path        string
		contentType string
	}

	type want struct {
		responseCode int
		request      string
		contentType  string
	}

	tests := []struct {
		name string
		set  set
		want want
	}{
		{
			name: "test #1 right response",
			set: set{
				method:      http.MethodPost,
				path:        "/api/shorten",
				contentType: "application/json",
			},
			want: want{
				responseCode: http.StatusCreated,
				request:      "",
				contentType:  "application/json",
			},
		},
		{
			name: "test #2 bad request",
			set: set{
				method:      http.MethodGet,
				path:        "/api/shorten",
				contentType: "",
			},
			want: want{
				responseCode: http.StatusBadRequest,
				request:      "",
				contentType:  "application/json",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := models.APIRequest{URL: "test"}
			data, _ := json.Marshal(body)
			req := httptest.NewRequest(test.set.method, test.set.path, bytes.NewBuffer(data))

			w := httptest.NewRecorder()
			h.apiShortenHandler(w, req)
			res := w.Result()

			defer res.Body.Close()

			_, err := io.ReadAll(res.Body)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			require.NoError(t, err)
		})
	}
}
