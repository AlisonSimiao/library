package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	userServiceURL := envOrDefault("USER_SERVICE_URL", "http://localhost:8081")
	bookServiceURL := envOrDefault("BOOK_SERVICE_URL", "http://localhost:8082")
	loanServiceURL := envOrDefault("LOAN_SERVICE_URL", "http://localhost:8083")

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Proxy para User Service
	r.Route("/api/users", func(r chi.Router) {
		r.HandleFunc("/*", proxyHandler(userServiceURL))
		r.HandleFunc("/", proxyHandler(userServiceURL))
	})

	// Proxy para Book Service
	r.Route("/api/books", func(r chi.Router) {
		r.HandleFunc("/*", proxyHandler(bookServiceURL))
		r.HandleFunc("/", proxyHandler(bookServiceURL))
	})

	// Proxy para Loan Service
	r.Route("/api/loans", func(r chi.Router) {
		r.HandleFunc("/*", proxyHandler(loanServiceURL))
		r.HandleFunc("/", proxyHandler(loanServiceURL))
	})

	log.Printf("[API Gateway] Rodando na porta %s", port)
	log.Printf("[API Gateway] User Service  -> %s", userServiceURL)
	log.Printf("[API Gateway] Book Service  -> %s", bookServiceURL)
	log.Printf("[API Gateway] Loan Service  -> %s", loanServiceURL)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Erro ao iniciar gateway: %v", err)
	}
}

func proxyHandler(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targetURL, err := url.Parse(target)
		if err != nil {
			http.Error(w, "Erro interno", http.StatusInternalServerError)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		proxy.ServeHTTP(w, r)
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
