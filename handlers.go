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
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	chirp, err := cfg.db.GetChirpByID(context.Background(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			sendErrorResponse(w, http.StatusNotFound, "Not found")
			return
		}
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendJSONResponse(w, http.StatusOK, chirp)
}

func (cfg *apiConfig) handleGetAllChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetAllChirps(context.Background())
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendJSONResponse(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	tokenString, err := getBearerToken(r)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, err := auth.ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var body struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(body.Body) > 140 {
		sendErrorResponse(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	chirp, err := cfg.db.CreateChirp(context.Background(), database.CreateChirpParams{
		UserID: userID,
		Body:   sanitizeChirpBody(body.Body),
	})
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendJSONResponse(w, http.StatusCreated, chirp)
}

func (cfg *apiConfig) handleRevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := getBearerToken(r)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	err = cfg.db.RevokeRefreshToken(context.Background(), refreshToken)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(204)
	w.Write([]byte{})
}

func (cfg *apiConfig) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := getBearerToken(r)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	user, err := cfg.db.GetUserFromRefreshToken(context.Background(), refreshToken)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, "Invalid token")
		return
	}
	tokenString, err := auth.MakeJWT(user.ID, cfg.jwtSecret, 1*time.Hour)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendJSONResponse(w, http.StatusOK, struct {
		Token string `json:"token"`
	}{Token: tokenString})
}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	user, err := cfg.db.GetUserByEmail(context.Background(), body.Email)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}
	ok, err := auth.VerifyPassword(body.Password, user.HashedPassword)
	if err != nil || !ok {
		sendErrorResponse(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}
	tokenString, err := auth.MakeJWT(user.ID, cfg.jwtSecret, 1*time.Hour)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	refreshTokenString, err := auth.MakeRefreshToken()
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	_, err = cfg.db.CreateRefreshToken(context.Background(), database.CreateRefreshTokenParams{
		Token:     refreshTokenString,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
	})
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendJSONResponse(w, http.StatusOK, struct {
		database.User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}{
		User:         user,
		Token:        tokenString,
		RefreshToken: refreshTokenString,
	})
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if !strings.Contains(body.Email, "@") {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid email")
		return
	}
	hashedPassword, err := auth.HashPassword(body.Password)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	user, err := cfg.db.CreateUser(context.Background(), database.CreateUserParams{
		Email:          body.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			sendErrorResponse(w, http.StatusConflict, "Email already exists")
			return
		}
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendJSONResponse(w, http.StatusCreated, user)
}

func (cfg *apiConfig) handleUpdateCredentials(w http.ResponseWriter, r *http.Request) {
	token, err := getBearerToken(r)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	hashedPassword, err := auth.HashPassword(body.Password)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	user, err := cfg.db.UpdateCredentials(context.Background(), database.UpdateCredentialsParams{
		ID:             userID,
		Email:          body.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendJSONResponse(w, http.StatusOK, user)
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
