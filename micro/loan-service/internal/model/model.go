package model

import "time"

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

type ErrorResponse struct {
	Error string `json:"error"`
}

// Structs para respostas dos outros serviços

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Book struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	Quantity int    `json:"quantity"`
}
