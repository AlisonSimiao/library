package main

import (
	"log"
	"net/http"

	"github.com/library-api/internal/config"
	"github.com/library-api/internal/database"
	"github.com/library-api/internal/handlers"
	"github.com/library-api/internal/repository"
	"github.com/library-api/internal/router"
)

func main() {
	// Carregar configuração
	cfg := config.Load()

	// Conectar ao banco de dados
	db := database.Connect(cfg.DatabaseURL)
	defer db.Close()

	// Executar migrations
	database.RunMigrations(db, "migrations")

	// Inicializar repositórios
	userRepo := repository.NewUserRepository(db)
	bookRepo := repository.NewBookRepository(db)
	loanRepo := repository.NewLoanRepository(db)

	// Inicializar handlers
	userHandler := handlers.NewUserHandler(userRepo)
	bookHandler := handlers.NewBookHandler(bookRepo)
	loanHandler := handlers.NewLoanHandler(loanRepo, bookRepo)

	// Configurar rotas
	r := router.NewRouter(userHandler, bookHandler, loanHandler)

	// Iniciar servidor
	log.Printf("Servidor rodando na porta %s", cfg.Port)
	log.Printf("API disponível em http://localhost:%s/api", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
