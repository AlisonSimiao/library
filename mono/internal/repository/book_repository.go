package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/library-api/internal/models"
)

type BookRepository struct {
	db *sqlx.DB
}

func NewBookRepository(db *sqlx.DB) *BookRepository {
	return &BookRepository{db: db}
}

func (r *BookRepository) Create(req models.CreateBookRequest) (*models.Book, error) {
	var book models.Book
	err := r.db.QueryRowx(
		"INSERT INTO books (title, author, isbn, quantity) VALUES ($1, $2, $3, $4) RETURNING id, title, author, isbn, quantity, created_at",
		req.Title, req.Author, req.ISBN, req.Quantity,
	).StructScan(&book)
	if err != nil {
		return nil, err
	}
	return &book, nil
}

func (r *BookRepository) FindAll() ([]models.Book, error) {
	var books []models.Book
	err := r.db.Select(&books, "SELECT id, title, author, isbn, quantity, created_at FROM books ORDER BY id")
	if err != nil {
		return nil, err
	}
	return books, nil
}

func (r *BookRepository) FindByID(id int) (*models.Book, error) {
	var book models.Book
	err := r.db.Get(&book, "SELECT id, title, author, isbn, quantity, created_at FROM books WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &book, nil
}

func (r *BookRepository) Search(query string) ([]models.Book, error) {
	var books []models.Book
	search := fmt.Sprintf("%%%s%%", query)
	err := r.db.Select(&books,
		"SELECT id, title, author, isbn, quantity, created_at FROM books WHERE LOWER(title) LIKE LOWER($1) OR LOWER(author) LIKE LOWER($1) ORDER BY id",
		search,
	)
	if err != nil {
		return nil, err
	}
	return books, nil
}
