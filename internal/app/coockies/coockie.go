package coockies

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

const (
	tokenExpire = time.Hour * 1
	tokenName   = "token"
	secretKey   = "verySecret1234"
)

func WithCoockies(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string
		userCookie, err := r.Cookie(tokenName)

		if err != nil {

			if r.RequestURI == "/api/user/urls" {
				http.Error(w, "empty cookie", http.StatusUnauthorized)
				return
			}

			token, err = createToken()

			if err != nil {
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
			userCookie = setCookie(w, token)
		}

		if _, err = getUID(userCookie.Value); err != nil {
			http.Error(w, "user id not found in cookie", http.StatusUnauthorized)
			return
		}

		if !isTokenValid(userCookie.Value) {
			token, err = createToken()

			if err != nil {
				http.Error(w, "", http.StatusInternalServerError)
				return
			}

			setCookie(w, token)
			userCookie = setCookie(w, token)
			ctx := context.WithValue(r.Context(), tokenName, userCookie.Value)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func createToken() (string, error) {
	uid := uuid.New().String()
	claims, err := createClaims(uid)

	if err != nil {
		return "", err
	}

	return claims, nil
}

type Claims struct {
	jwt.RegisteredClaims
	uid string
}

func createClaims(uid string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpire)),
		},
		uid: uid,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func setCookie(w http.ResponseWriter, token string) *http.Cookie {
	cookie := &http.Cookie{
		Name:     tokenName,
		Value:    token,
		HttpOnly: true,
	}

	http.SetCookie(w, cookie)
	return cookie
}

func getUID(token string) (string, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})
	if err != nil {
		return "", err
	}

	return claims.uid, nil
}

func isTokenValid(token string) bool {
	claims := &Claims{}

	t, err := jwt.ParseWithClaims(token, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})
	if err != nil {
		return false
	}

	if !t.Valid {
		return false
	}

	return true
}
