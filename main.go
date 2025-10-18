package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/aleksaelezovic/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
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

func (cfg *apiConfig) metricsResetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("Metrics reset successfully."))
}

var profaneWords = []string{"kerfuffle", "sharbert", "fornax"}

func main() {
	godotenv.Load()
	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	cfg := &apiConfig{db: database.New(db)}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})
	fsHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", cfg.middlewareMetricsInc(fsHandler))
	mux.HandleFunc("GET /admin/metrics", cfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", cfg.metricsResetHandler)
	mux.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, r *http.Request) {
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
	})

	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
	server.ListenAndServe()
}
