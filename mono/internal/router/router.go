package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/library-api/internal/handlers"
)

func NewRouter(userHandler *handlers.UserHandler, bookHandler *handlers.BookHandler, loanHandler *handlers.LoanHandler) *chi.Mux {
	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	// Rotas
	r.Route("/api", func(r chi.Router) {
		// Usuários
		r.Route("/users", func(r chi.Router) {
			r.Post("/", userHandler.Create)
			r.Get("/", userHandler.FindAll)
			r.Get("/{id}", userHandler.FindByID)
		})

		// Livros
		r.Route("/books", func(r chi.Router) {
			r.Post("/", bookHandler.Create)
			r.Get("/", bookHandler.FindAll)
			r.Get("/{id}", bookHandler.FindByID)
		})

		// Empréstimos
		r.Route("/loans", func(r chi.Router) {
			r.Post("/", loanHandler.Create)
			r.Get("/", loanHandler.FindAll)
			r.Patch("/{id}/return", loanHandler.Return)
		})
	})

	return r
}
