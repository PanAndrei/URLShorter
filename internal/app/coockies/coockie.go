package coockies

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type Token string

const (
	tokenExpire       = time.Hour * 1
	TokenName   Token = "token"
	secretKey         = "verySecret1234"
)

func WithCoockies(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string
		userCookie, err := r.Cookie(string(TokenName))

		if err != nil {

			if r.RequestURI == "/api/user/urls" {
				http.Error(w, "empty cookie", http.StatusUnauthorized)
				return
			}
			println("tyt1")
			token, err = createToken()

			if err != nil {
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
			userCookie = setCookie(w, token)
		}

		if _, err = GetUID(userCookie.Value); err != nil {
			http.Error(w, "user id not found in cookie", http.StatusUnauthorized)
			return
		}

		if !isTokenValid(userCookie.Value) {
			token, err = createToken()
			println("tyt2")
			if err != nil {
				http.Error(w, "", http.StatusInternalServerError)
				return
			}

			setCookie(w, token)
			userCookie = setCookie(w, token)
		}

		ctx := context.WithValue(r.Context(), TokenName, userCookie.Value)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func createToken() (string, error) {
	uid := uuid.New().String()
	println(uid, "gg5")
	claims, err := createClaims(uid)

	if err != nil {
		return "", err
	}

	println(claims, "gg1")
	return claims, nil
}

type Claims struct {
	jwt.RegisteredClaims
	UID string
}

func createClaims(uid string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpire)),
		},
		UID: uid,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	println(tokenString, "gg2")

	return tokenString, nil
}

func setCookie(w http.ResponseWriter, token string) *http.Cookie {
	cookie := &http.Cookie{
		Name:     token,
		Value:    token,
		MaxAge:   10000,
		Path:     "/",
		HttpOnly: true,
	}

	http.SetCookie(w, cookie)
	return cookie
}

func GetUID(token string) (string, error) {
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

	println(claims.UID, "gg3")

	return claims.UID, nil
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

	println(t.Valid, "gg4")
	return t.Valid
}
