package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aleksaelezovic/chirpy/internal/auth"
	"github.com/aleksaelezovic/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handleGetChirpByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(400)
		w.Write(fmt.Appendf(make([]byte, 0), "{\"error\": \"%s\"}", err.Error()))
		return
	}
	chirp, err := cfg.db.GetChirpByID(context.Background(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(404)
			w.Write([]byte("{\"error\": \"not found\"}"))
			return
		}
		w.WriteHeader(500)
		w.Write(fmt.Appendf(make([]byte, 0), "{\"error\": \"%s\"}", err.Error()))
		return
	}
	chirpJson, err := json.Marshal(chirp)
	if err != nil {
		w.WriteHeader(500)
		w.Write(fmt.Appendf(make([]byte, 0), "{\"error\": \"%s\"}", err.Error()))
		return
	}
	w.WriteHeader(200)
	w.Write(chirpJson)
}

func (cfg *apiConfig) handleGetAllChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetAllChirps(context.Background())
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		w.WriteHeader(500)
		w.Write(fmt.Appendf(make([]byte, 0), "{\"error\": \"%s\"}", err.Error()))
		return
	}
	chirpsJson, err := json.Marshal(chirps)
	if err != nil {
		w.WriteHeader(500)
		w.Write(fmt.Appendf(make([]byte, 0), "{\"error\": \"%s\"}", err.Error()))
		return
	}
	w.WriteHeader(200)
	w.Write(chirpsJson)
}

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	tokenString, err := getBearerToken(r)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte("{\"error\": \"Unauthorized\"}"))
		return
	}
	userID, err := auth.ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte("{\"error\": \"Unauthorized\"}"))
		return
	}

	var body struct {
		Body string `json:"body"`
	}
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
	sanitizedBody := strings.Join(newWords, " ")

	chirp, err := cfg.db.CreateChirp(context.Background(), database.CreateChirpParams{
		UserID: userID,
		Body:   sanitizedBody,
	})
	if err != nil {
		w.WriteHeader(500)
		w.Write(fmt.Appendf(make([]byte, 0), "{\"error\": \"%s\"}", err.Error()))
		return
	}
	chirpJson, err := json.Marshal(chirp)
	if err != nil {
		w.WriteHeader(500)
		w.Write(fmt.Appendf(make([]byte, 0), "{\"error\": \"%s\"}", err.Error()))
		return
	}
	w.WriteHeader(201)
	w.Write(chirpJson)
}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		ExpiresIn int    `json:"expires_in_seconds"`
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(400)
		w.Write(fmt.Appendf(make([]byte, 0), "{\"error\": \"%s\"}", err.Error()))
		return
	}
	user, err := cfg.db.GetUserByEmail(context.Background(), body.Email)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte("{\"error\": \"Incorrect email or password\"}"))
		return
	}
	ok, err := auth.VerifyPassword(body.Password, user.HashedPassword)
	if err != nil || !ok {
		w.WriteHeader(401)
		w.Write([]byte("{\"error\": \"Incorrect email or password\"}"))
		return
	}
	if body.ExpiresIn == 0 {
		body.ExpiresIn = 3600 // Default expiration time in seconds = 1 hour
	}
	tokenString, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Duration(body.ExpiresIn)*time.Second)
	if err != nil {
		w.WriteHeader(500)
		w.Write(fmt.Appendf(make([]byte, 0), "{\"error\": \"%s\"}", err.Error()))
		return
	}
	data, err := json.Marshal(struct {
		database.User
		Token string `json:"token"`
	}{
		User:  user,
		Token: tokenString,
	})
	if err != nil {
		w.WriteHeader(500)
		w.Write(fmt.Appendf(make([]byte, 0), "{\"error\": \"%s\"}", err.Error()))
		return
	}
	w.WriteHeader(200)
	w.Write(data)
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
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
	hashedPassword, err := auth.HashPassword(body.Password)
	if err != nil {
		w.WriteHeader(500)
		w.Write(fmt.Appendf(make([]byte, 0), "{\"error\": \"%s\"}", err.Error()))
		return
	}
	user, err := cfg.db.CreateUser(context.Background(), database.CreateUserParams{
		Email:          body.Email,
		HashedPassword: hashedPassword,
	})
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
