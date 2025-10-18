package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var profaneWords = []string{"kerfuffle", "sharbert", "fornax"}

func getBearerToken(r *http.Request) (string, error) {
	token := r.Header.Get("Authorization")
	if len(token) > 7 && strings.ToUpper(token[:7]) == "BEARER " {
		return token[7:], nil
	}
	return "", errors.New("invalid token")
}

func sendJSONResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
	} else {
		w.Write(jsonBytes)
	}
}

func sendErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(fmt.Appendf(make([]byte, 0, len(message)+13), "{\"error\": \"%s\"}", message))
}

func sanitizeChirpBody(body string) string {
	oldWords := strings.Split(body, " ")
	newWords := make([]string, len(oldWords))
	for i, word := range oldWords {
		for _, profaneWord := range profaneWords {
			if strings.EqualFold(strings.ToLower(word), strings.ToLower(profaneWord)) {
				newWords[i] = "****"
				break
			} else {
				newWords[i] = word
			}
		}
	}
	return strings.Join(newWords, " ")
}
