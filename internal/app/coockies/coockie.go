package coockies

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type AuthToken string

const (
	tokenExpire           = time.Hour * 1
	TokenName   AuthToken = "auth_token"
	secretKey             = "verySecret1234"
)

func WithCoockies(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var cookieString string
		token, err := r.Cookie(string(TokenName))

		if err != nil {
			cookieString, err = createToken()
			if err != nil {
				http.Error(w, "failed to generate a new token", http.StatusInternalServerError)
				return
			}
			setCookie(w, cookieString, r)
		} else {
			_, err := GetUID(token.Value)
			if err != nil {
				// cookieString, err = createToken()
				// if err != nil {
				// 	http.Error(w, "failed to generate a new token", http.StatusInternalServerError)
				// 	return
				// }
				// setCookie(w, cookieString, r)
				http.Error(w, "user id not found", http.StatusUnauthorized)
				return
			}

			valid, err := isTokenValid(token.Value)
			if err != nil || !valid {
				// cookieString, err = createToken()
				// if err != nil {
				// 	http.Error(w, "failed to generate a new token", http.StatusInternalServerError)
				// 	return
				// }
				// setCookie(w, cookieString, r)
				http.Error(w, "user id not found", http.StatusUnauthorized)
				return
			}
			cookieString = token.Value
		}

		ctx := context.WithValue(r.Context(), TokenName, cookieString)
		next.ServeHTTP(w, r.WithContext(ctx))
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

	return tokenString, nil
}

func setCookie(w http.ResponseWriter, token string, r *http.Request) *http.Cookie {
	cookie := &http.Cookie{
		Name:     string(TokenName),
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
				return "", fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})
	if err != nil {
		return "", err
	}
	if claims.UID == "" {
		return "", fmt.Errorf("uid is empty")
	}

	return claims.UID, nil
}

func isTokenValid(token string) (bool, error) {
	claims := &Claims{}

	t, err := jwt.ParseWithClaims(token, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return false, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})
	if err != nil {
		return false, err
	}

	return t.Valid, nil
}
