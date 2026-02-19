package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/library-api/internal/models"
	"github.com/library-api/internal/repository"
)

type LoanHandler struct {
	loanRepo *repository.LoanRepository
	bookRepo *repository.BookRepository
}

func NewLoanHandler(loanRepo *repository.LoanRepository, bookRepo *repository.BookRepository) *LoanHandler {
	return &LoanHandler{
		loanRepo: loanRepo,
		bookRepo: bookRepo,
	}
}

func (h *LoanHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateLoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "JSON inválido"})
		return
	}

	if req.UserID == 0 || req.BookID == 0 {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "user_id e book_id são obrigatórios"})
		return
	}

	// Verificar se o livro existe e tem exemplares disponíveis
	book, err := h.bookRepo.FindByID(req.BookID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "Livro não encontrado"})
		return
	}

	activeLoans, err := h.loanRepo.CountActiveByBookID(req.BookID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "Erro ao verificar disponibilidade"})
		return
	}

	if activeLoans >= book.Quantity {
		writeJSON(w, http.StatusConflict, models.ErrorResponse{Error: "Não há exemplares disponíveis deste livro"})
		return
	}

	loan, err := h.loanRepo.Create(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "Erro ao registrar empréstimo: " + err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, loan)
}

func (h *LoanHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")

	var loans []models.Loan
	var err error

	if userIDStr != "" {
		userID, parseErr := strconv.Atoi(userIDStr)
		if parseErr != nil {
			writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "user_id inválido"})
			return
		}
		loans, err = h.loanRepo.FindByUserID(userID)
	} else {
		loans, err = h.loanRepo.FindAll()
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "Erro ao buscar empréstimos"})
		return
	}

	writeJSON(w, http.StatusOK, loans)
}

func (h *LoanHandler) Return(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "ID inválido"})
		return
	}

	loan, err := h.loanRepo.Return(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "Empréstimo não encontrado ou já devolvido"})
		return
	}

	writeJSON(w, http.StatusOK, loan)
}
