package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/user-service/internal/database"
	"github.com/user-service/internal/handler"
	"github.com/user-service/internal/repository"
)

func main() {
	godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	db := database.Connect(os.Getenv("DATABASE_URL"))
	defer db.Close()

	database.RunMigrations(db, "migrations")

	repo := repository.New(db)
	h := handler.New(repo)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/users", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.FindAll)
		r.Get("/{id}", h.FindByID)
	})

	log.Printf("[User Service] Rodando na porta %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
