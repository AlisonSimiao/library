package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/user-service/internal/model"
)

type UserRepository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(req model.CreateUserRequest) (*model.User, error) {
	var user model.User
	err := r.db.QueryRowx(
		"INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, name, email, created_at",
		req.Name, req.Email,
	).StructScan(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindAll() ([]model.User, error) {
	var users []model.User
	err := r.db.Select(&users, "SELECT id, name, email, created_at FROM users ORDER BY id")
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) FindByID(id int) (*model.User, error) {
	var user model.User
	err := r.db.Get(&user, "SELECT id, name, email, created_at FROM users WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
