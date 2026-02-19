package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/library-api/internal/models"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(req models.CreateUserRequest) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowx(
		"INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, name, email, created_at",
		req.Name, req.Email,
	).StructScan(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindAll() ([]models.User, error) {
	var users []models.User
	err := r.db.Select(&users, "SELECT id, name, email, created_at FROM users ORDER BY id")
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) FindByID(id int) (*models.User, error) {
	var user models.User
	err := r.db.Get(&user, "SELECT id, name, email, created_at FROM users WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
