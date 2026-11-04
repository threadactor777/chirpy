package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

// Holds stateful, in-memory data to keep track
type apiConfig struct {
	fileserverHits atomic.Int32
}

// Middleware to safely increment the server hits
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// Return the number of hits
func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// hits := fmt.Sprintf("Hits: %v", cfg.fileserverHits.Load())
	hits := fmt.Sprintf(
		`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %v times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())

	w.Write([]byte(hits))
}

// Reset to zero
func (cfg *apiConfig) handlerZero(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	hits := fmt.Sprintf("Hits: %v", cfg.fileserverHits.Load())
	w.Write([]byte(hits))
}

// Validate Chirp Length
func (cfg *apiConfig) handlerValidate(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chrip is too long")
		return
	}

	cleaned_up := wordCheck(params.Body)

	type validResponse struct {
		Cleaned string `json:"cleaned_body"`
	}

	respondWithJSON(w, http.StatusOK, validResponse{
		Cleaned: cleaned_up,
	})

}

// Check for Bad Words
func wordCheck(message string) string {
	split := strings.Split(message, " ")
	for index, word := range split {
		switch strings.ToLower(word) {
		case "kerfuffle":
			split[index] = "****"
		case "sharbert":
			split[index] = "****"
		case "fornax":
			split[index] = "****"
		default:
			continue
		}
	}

	joined := strings.Join(split, " ")

	return joined
}

// Respond with JSON
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error Marshallin JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

// Error Response
func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func main() {

	apiCfg := &apiConfig{}

	// Routing requests
	mux := http.NewServeMux()

	// Add handler to root path to serve index.html
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	// Readiness Endpoint
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Metrics Endpoint
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)

	// Reset Endpoint
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerZero)

	// Validate chirp length endpoint
	mux.HandleFunc("POST /api/validate_chirp", apiCfg.handlerValidate)

	// Server Struct
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}

}
