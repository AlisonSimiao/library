package model

import "time"

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

type ErrorResponse struct {
	Error string `json:"error"`
}
