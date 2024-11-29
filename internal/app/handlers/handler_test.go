package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	repo "URLShorter/internal/app/repository"

	assert "github.com/stretchr/testify/assert"
)

type MockShortener struct{}

func (m *MockShortener) SetShortURL(u *repo.URL) *repo.URL {
	u.ShortURL = "m.ShortURL"

	return u
}

func (m *MockShortener) GetFullURL(u *repo.URL) (*repo.URL, error) {

	return &repo.URL{FullUrl: "m.FullURL"}, nil
}

func TestMainPostHandler(t *testing.T) {
	h := NewHandlers(&MockShortener{})

	type want struct {
		responseCode int
		request      string
		contentType  string
	}

	tests := []struct {
		name string
		want want
	}{
		{
			name: "test #1",
			want: want{
				responseCode: http.StatusCreated,
				request:      "",
				contentType:  "text/plain",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			w := httptest.NewRecorder()
			h.mainPostHandler(w, req)
			res := w.Result()

			assert.Equal(t, test.want.responseCode, res.StatusCode)
		})
	}

}
