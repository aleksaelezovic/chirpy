package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(400)
		w.Write(fmt.Appendf(make([]byte, 0), "{\"error\": \"%s\"}", err.Error()))
		return
	}
	if !strings.Contains(body.Email, "@") {
		w.WriteHeader(400)
		w.Write([]byte("{\"error\": \"Invalid email\"}"))
		return
	}
	user, err := cfg.db.CreateUser(context.Background(), body.Email)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			w.WriteHeader(409)
			w.Write([]byte("{\"error\": \"Email already exists\"}"))
			return
		}
		w.WriteHeader(500)
		w.Write(fmt.Appendf(make([]byte, 0), "{\"error\": \"%s\"}", err.Error()))
		return
	}
	userJson, err := json.Marshal(user)
	if err != nil {
		w.WriteHeader(500)
		w.Write(fmt.Appendf(make([]byte, 0), "{\"error\": \"%s\"}", err.Error()))
		return
	}
	w.WriteHeader(201)
	w.Write(userJson)
}

func handleChirpValidation(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Body string `json:"body"`
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(400)
		w.Write(fmt.Appendf(make([]byte, 0), "{\"error\": \"%s\"}", err.Error()))
		return
	}
	if len(body.Body) > 140 {
		w.WriteHeader(400)
		w.Write([]byte("{\"error\": \"Chirp is too long\"}"))
		return
	}
	oldWords := strings.Split(body.Body, " ")
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
	sanitized := strings.Join(newWords, " ")
	w.WriteHeader(200)
	w.Write(fmt.Appendf(make([]byte, 0), "{\"cleaned_body\": \"%s\"}", sanitized))
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write(fmt.Appendf(make([]byte, 0), `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load()))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	if !cfg.isDev {
		w.WriteHeader(403)
		w.Write([]byte("Forbidden"))
		return
	}
	cfg.fileserverHits.Store(0)
	err := cfg.db.DeleteAllUsers(context.Background())
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Internal Server Error"))
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("Reset successfully."))
}
