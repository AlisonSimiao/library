package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/loan-service/internal/client"
	"github.com/loan-service/internal/model"
	"github.com/loan-service/internal/repository"
)

type LoanHandler struct {
	repo          *repository.LoanRepository
	serviceClient *client.ServiceClient
}

func New(repo *repository.LoanRepository, serviceClient *client.ServiceClient) *LoanHandler {
	return &LoanHandler{
		repo:          repo,
		serviceClient: serviceClient,
	}
}

func (h *LoanHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateLoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "JSON inválido"})
		return
	}
	if req.UserID == 0 || req.BookID == 0 {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "user_id e book_id são obrigatórios"})
		return
	}

	// Validar usuário via User Service (HTTP)
	_, err := h.serviceClient.GetUser(req.UserID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "Usuário não encontrado: " + err.Error()})
		return
	}

	// Validar livro via Book Service (HTTP)
	book, err := h.serviceClient.GetBook(req.BookID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "Livro não encontrado: " + err.Error()})
		return
	}

	// Verificar disponibilidade
	activeLoans, err := h.repo.CountActiveByBookID(req.BookID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "Erro ao verificar disponibilidade"})
		return
	}
	if activeLoans >= book.Quantity {
		writeJSON(w, http.StatusConflict, model.ErrorResponse{Error: "Não há exemplares disponíveis deste livro"})
		return
	}

	loan, err := h.repo.Create(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "Erro ao registrar empréstimo: " + err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, loan)
}

func (h *LoanHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	var loans []model.Loan
	var err error
	if userIDStr != "" {
		userID, parseErr := strconv.Atoi(userIDStr)
		if parseErr != nil {
			writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "user_id inválido"})
			return
		}
		loans, err = h.repo.FindByUserID(userID)
	} else {
		loans, err = h.repo.FindAll()
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "Erro ao buscar empréstimos"})
		return
	}
	writeJSON(w, http.StatusOK, loans)
}

func (h *LoanHandler) Return(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "ID inválido"})
		return
	}
	loan, err := h.repo.Return(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "Empréstimo não encontrado ou já devolvido"})
		return
	}
	writeJSON(w, http.StatusOK, loan)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
