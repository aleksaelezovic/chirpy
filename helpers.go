package main

import (
	"errors"
	"net/http"
	"strings"
)

func getBearerToken(r *http.Request) (string, error) {
	token := r.Header.Get("Authorization")
	if len(token) > 7 && strings.ToUpper(token[:7]) == "BEARER " {
		return token[7:], nil
	}
	return "", errors.New("invalid token")
}
