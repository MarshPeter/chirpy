package main

import (
	"fmt"
	"net/http"
)

type apiConfig struct {
	fileserverhits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverhits++

		next.ServeHTTP(w, r)
	})
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	// set content type header
	w.Header().Set("Content-type", "text/plain; charset=utf-8")
	// Write the status code
	w.WriteHeader(http.StatusOK)
	// Write the body
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) incrementCounterHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("Hits: %v", cfg.fileserverhits)))
}

func (cfg *apiConfig) resetCounterHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverhits = 0
	w.Write([]byte("Reset done\n"))
}

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", readinessHandler)
	apiCfg := apiConfig{}
	fileServer := http.FileServer(http.Dir("."))
	mux.Handle("GET /app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(fileServer)))
	mux.HandleFunc("GET /metrics", apiCfg.incrementCounterHandler)
	mux.HandleFunc("GET /reset", apiCfg.resetCounterHandler)

	http.ListenAndServe(":8080", mux)
}
