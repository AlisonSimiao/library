package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/library-api/internal/models"
	"github.com/library-api/internal/repository"
)

type UserHandler struct {
	repo *repository.UserRepository
}

func NewUserHandler(repo *repository.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "JSON inválido"})
		return
	}

	if req.Name == "" || req.Email == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "Nome e email são obrigatórios"})
		return
	}

	user, err := h.repo.Create(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "Erro ao criar usuário: " + err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h *UserHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	users, err := h.repo.FindAll()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "Erro ao buscar usuários"})
		return
	}

	writeJSON(w, http.StatusOK, users)
}

func (h *UserHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "ID inválido"})
		return
	}

	user, err := h.repo.FindByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "Usuário não encontrado"})
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
