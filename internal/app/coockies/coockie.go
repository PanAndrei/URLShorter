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
		fmt.Println("Middleware - token: ", token)
		if err != nil {
			fmt.Println("Middleware - no cookie found")
			cookieString, err = createToken()
			if err != nil {
				http.Error(w, "failed to generate a new token", http.StatusInternalServerError)
				return
			}
			fmt.Println("Middleware - generated a new token: ", cookieString)
			setCookie(w, cookieString, r)
		} else if _, err := GetUID(token.Value); err != nil {
			fmt.Println("Middleware - user id not found in cookie")
			http.Error(w, "user id not found", http.StatusUnauthorized)
			return
		} else if !isTokenValid(token.Value) {
			fmt.Println("Middleware - token invalid")
			cookieString, err = createToken()
			if err != nil {
				http.Error(w, "failed to generate a new token", http.StatusInternalServerError)
				return
			}
			fmt.Println("Middleware - re-generated a new token: ", cookieString)
			setCookie(w, cookieString, r)
		} else {
			cookieString = token.Value
			fmt.Println("Middleware - valid token found in cookie")
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
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})
	if err != nil {
		return "", err
	}

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

	return t.Valid
}
