package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type apiConfig struct {
	fileserverhits int
}

type Chirp struct {
	ID int `json:"id"`
	Body string `json:"body"`
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

func validateChirpyHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type cleanedReturn struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err :=  decoder.Decode(&params)

	if err != nil {
		responseWithError(w, 500, "something went wrong")
	}

	if len(params.Body) > 140 {
		responseWithError(w, 400, "Chirp is too long")
		return
	}

	cleanedBody := cleanBody(params.Body)

	respBody := cleanedReturn {
		CleanedBody: cleanedBody,
	}

	responseWithJson(w, 200, respBody)
}	

func cleanBody(body string) string {
	if len(body) == 0 {
		return ""
	}
	badWords := []string {
		"kerfuffle",
		"sharbert",
		"fornax",
	}

	words := strings.Split(body, " ") 

	for i, word := range words {
		for _, badWord := range badWords {
			if strings.ToLower(word) == badWord {
				words[i] = "****"
			}
		}
	}

	return strings.Join(words, " ")
}

func responseWithError(w http.ResponseWriter, code int, msg string) {
	type errorReturnValues struct {
		Error string `json:"error"`
	}
	w.WriteHeader(code)
	respBody := errorReturnValues {
		Error: msg,
	}
	data, _ := json.Marshal(respBody)
	w.Write(data)
}

func responseWithJson(w http.ResponseWriter, code int, payload interface{}) {
	dat, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
	w.Write(dat)
} 

func (cfg *apiConfig) adminMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	html := fmt.Sprintf(`
		<html>

			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>

		</html>
	`, cfg.fileserverhits)
	w.Write([]byte(html))
}

func (cfg *apiConfig) incrementCounterHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("Hits: %v", cfg.fileserverhits)))
}

func (cfg *apiConfig) resetCounterHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverhits = 0
	w.Write([]byte("Reset done\n"))
}

func main() {
	
	db, err := database.NewDB("database.json")
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/healthz", readinessHandler)
	apiCfg := apiConfig{}
	fileServer := http.FileServer(http.Dir("."))
	mux.Handle("GET /app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(fileServer)))
	mux.HandleFunc("GET /api/metrics", apiCfg.incrementCounterHandler)
	mux.HandleFunc("GET /api/reset", apiCfg.resetCounterHandler)
	mux.HandleFunc("POST /api/validate_chirp", validateChirpyHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.adminMetrics)

	http.ListenAndServe(":8080", mux)
}
