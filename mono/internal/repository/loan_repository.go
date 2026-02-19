package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/library-api/internal/models"
)

type LoanRepository struct {
	db *sqlx.DB
}

func NewLoanRepository(db *sqlx.DB) *LoanRepository {
	return &LoanRepository{db: db}
}

func (r *LoanRepository) Create(req models.CreateLoanRequest) (*models.Loan, error) {
	var loan models.Loan
	err := r.db.QueryRowx(
		"INSERT INTO loans (user_id, book_id) VALUES ($1, $2) RETURNING id, user_id, book_id, loaned_at, returned_at, status",
		req.UserID, req.BookID,
	).StructScan(&loan)
	if err != nil {
		return nil, err
	}
	return &loan, nil
}

func (r *LoanRepository) FindAll() ([]models.Loan, error) {
	var loans []models.Loan
	err := r.db.Select(&loans, "SELECT id, user_id, book_id, loaned_at, returned_at, status FROM loans ORDER BY id")
	if err != nil {
		return nil, err
	}
	return loans, nil
}

func (r *LoanRepository) FindByUserID(userID int) ([]models.Loan, error) {
	var loans []models.Loan
	err := r.db.Select(&loans,
		"SELECT id, user_id, book_id, loaned_at, returned_at, status FROM loans WHERE user_id = $1 ORDER BY id",
		userID,
	)
	if err != nil {
		return nil, err
	}
	return loans, nil
}

func (r *LoanRepository) Return(id int) (*models.Loan, error) {
	var loan models.Loan
	err := r.db.QueryRowx(
		"UPDATE loans SET status = 'returned', returned_at = NOW() WHERE id = $1 AND status = 'active' RETURNING id, user_id, book_id, loaned_at, returned_at, status",
		id,
	).StructScan(&loan)
	if err != nil {
		return nil, err
	}
	return &loan, nil
}

func (r *LoanRepository) CountActiveByBookID(bookID int) (int, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM loans WHERE book_id = $1 AND status = 'active'", bookID)
	if err != nil {
		return 0, err
	}
	return count, nil
}
