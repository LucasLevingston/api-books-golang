package book

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("book not found")

type Book struct {
	ID          int64     `db:"id"          json:"id"`
	Name        string    `db:"name"        json:"name"`
	Author      string    `db:"author"      json:"author"`
	Year        int       `db:"year"        json:"year"`
	Masterpiece bool      `db:"masterpiece" json:"masterpiece"`
	CreatedAt   time.Time `db:"created_at"  json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"  json:"updated_at"`
}

type ListFilter struct {
	Author      string
	Masterpiece *bool
	Year        *int
	Page        int
	PageSize    int
	SortBy      string
	SortDir     string
}

type CreateInput struct {
	Name        string `json:"name"        validate:"required,min=1,max=255"`
	Author      string `json:"author"      validate:"required,min=1,max=255"`
	Year        int    `json:"year"        validate:"required,min=1,max=2100"`
	Masterpiece bool   `json:"masterpiece"`
}

type UpdateInput struct {
	Name        *string `json:"name"        validate:"omitempty,min=1,max=255"`
	Author      *string `json:"author"      validate:"omitempty,min=1,max=255"`
	Year        *int    `json:"year"        validate:"omitempty,min=1,max=2100"`
	Masterpiece *bool   `json:"masterpiece"`
}

type PageResult struct {
	Data       []Book `json:"data"`
	Total      int    `json:"total"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	TotalPages int    `json:"total_pages"`
}

type Repository interface {
	FindAll(filter ListFilter) (PageResult, error)
	FindByID(id int64) (Book, error)
	Create(input CreateInput) (Book, error)
	Update(id int64, input UpdateInput) (Book, error)
	Delete(id int64) error
}

type Service interface {
	List(filter ListFilter) (PageResult, error)
	Get(id int64) (Book, error)
	Create(input CreateInput) (Book, error)
	Update(id int64, input UpdateInput) (Book, error)
	Delete(id int64) error
}
