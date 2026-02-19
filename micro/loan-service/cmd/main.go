package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/loan-service/internal/client"
	"github.com/loan-service/internal/database"
	"github.com/loan-service/internal/handler"
	"github.com/loan-service/internal/repository"
)

func main() {
	godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	db := database.Connect(os.Getenv("DATABASE_URL"))
	defer db.Close()

	database.RunMigrations(db, "migrations")

	repo := repository.New(db)
	serviceClient := client.New(
		os.Getenv("USER_SERVICE_URL"),
		os.Getenv("BOOK_SERVICE_URL"),
	)
	h := handler.New(repo, serviceClient)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/loans", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.FindAll)
		r.Patch("/{id}/return", h.Return)
	})

	log.Printf("[Loan Service] Rodando na porta %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
