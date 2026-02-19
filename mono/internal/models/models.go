package models

import "time"

// ==================== User ====================

type User struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ==================== Book ====================

type Book struct {
	ID        int       `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Author    string    `json:"author" db:"author"`
	ISBN      string    `json:"isbn" db:"isbn"`
	Quantity  int       `json:"quantity" db:"quantity"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateBookRequest struct {
	Title    string `json:"title"`
	Author   string `json:"author"`
	ISBN     string `json:"isbn"`
	Quantity int    `json:"quantity"`
}

// ==================== Loan ====================

type Loan struct {
	ID         int        `json:"id" db:"id"`
	UserID     int        `json:"user_id" db:"user_id"`
	BookID     int        `json:"book_id" db:"book_id"`
	LoanedAt   time.Time  `json:"loaned_at" db:"loaned_at"`
	ReturnedAt *time.Time `json:"returned_at,omitempty" db:"returned_at"`
	Status     string     `json:"status" db:"status"`
}

type CreateLoanRequest struct {
	UserID int `json:"user_id"`
	BookID int `json:"book_id"`
}

// ==================== API Response ====================

type ErrorResponse struct {
	Error string `json:"error"`
}
