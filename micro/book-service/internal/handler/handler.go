package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/book-service/internal/model"
	"github.com/book-service/internal/repository"
	"github.com/go-chi/chi/v5"
)

type BookHandler struct {
	repo *repository.BookRepository
}

func New(repo *repository.BookRepository) *BookHandler {
	return &BookHandler{repo: repo}
}

func (h *BookHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "JSON inválido"})
		return
	}
	if req.Title == "" || req.Author == "" {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "Título e autor são obrigatórios"})
		return
	}
	if req.Quantity <= 0 {
		req.Quantity = 1
	}
	book, err := h.repo.Create(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "Erro ao criar livro: " + err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, book)
}

func (h *BookHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	var books []model.Book
	var err error
	if query != "" {
		books, err = h.repo.Search(query)
	} else {
		books, err = h.repo.FindAll()
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "Erro ao buscar livros"})
		return
	}
	writeJSON(w, http.StatusOK, books)
}

func (h *BookHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "ID inválido"})
		return
	}
	book, err := h.repo.FindByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "Livro não encontrado"})
		return
	}
	writeJSON(w, http.StatusOK, book)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
