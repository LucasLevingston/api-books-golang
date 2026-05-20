package book

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

var allowedSortCols = map[string]bool{
	"name": true, "author": true, "masterpiece": true, "year": true,
}

type postgresRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) FindAll(f ListFilter) (PageResult, error) {
	var conds []string
	var args []interface{}
	idx := 1

	if f.Author != "" {
		conds = append(conds, fmt.Sprintf("author ILIKE $%d", idx))
		args = append(args, "%"+f.Author+"%")
		idx++
	}
	if f.Masterpiece != nil {
		conds = append(conds, fmt.Sprintf("masterpiece = $%d", idx))
		args = append(args, *f.Masterpiece)
		idx++
	}
	if f.Year != nil {
		conds = append(conds, fmt.Sprintf("year = $%d", idx))
		args = append(args, *f.Year)
		idx++
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	var total int
	if err := r.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM books %s", where), args...).Scan(&total); err != nil {
		return PageResult{}, fmt.Errorf("count books: %w", err)
	}

	sortCol := safeSortCol(f.SortBy)
	sortDir := safeSortDir(f.SortDir)

	offset := (f.Page - 1) * f.PageSize
	query := fmt.Sprintf(
		`SELECT id, name, author, year, masterpiece, created_at, updated_at
		 FROM books %s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		where, sortCol, sortDir, idx, idx+1,
	)
	args = append(args, f.PageSize, offset)

	var books []Book
	if err := r.db.Select(&books, query, args...); err != nil {
		return PageResult{}, fmt.Errorf("select books: %w", err)
	}
	if books == nil {
		books = []Book{}
	}

	totalPages := total / f.PageSize
	if total%f.PageSize != 0 {
		totalPages++
	}

	return PageResult{
		Data:       books,
		Total:      total,
		Page:       f.Page,
		PageSize:   f.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *postgresRepository) FindByID(id int64) (Book, error) {
	var b Book
	err := r.db.Get(&b,
		`SELECT id, name, author, year, masterpiece, created_at, updated_at FROM books WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Book{}, ErrNotFound
	}
	if err != nil {
		return Book{}, fmt.Errorf("find book: %w", err)
	}
	return b, nil
}

func (r *postgresRepository) Create(input CreateInput) (Book, error) {
	var b Book
	err := r.db.QueryRowx(
		`INSERT INTO books (name, author, year, masterpiece)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, name, author, year, masterpiece, created_at, updated_at`,
		input.Name, input.Author, input.Year, input.Masterpiece,
	).StructScan(&b)
	if err != nil {
		return Book{}, fmt.Errorf("create book: %w", err)
	}
	return b, nil
}

func (r *postgresRepository) Update(id int64, input UpdateInput) (Book, error) {
	var sets []string
	var args []interface{}
	idx := 1

	if input.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", idx))
		args = append(args, *input.Name)
		idx++
	}
	if input.Author != nil {
		sets = append(sets, fmt.Sprintf("author = $%d", idx))
		args = append(args, *input.Author)
		idx++
	}
	if input.Year != nil {
		sets = append(sets, fmt.Sprintf("year = $%d", idx))
		args = append(args, *input.Year)
		idx++
	}
	if input.Masterpiece != nil {
		sets = append(sets, fmt.Sprintf("masterpiece = $%d", idx))
		args = append(args, *input.Masterpiece)
		idx++
	}
	if len(sets) == 0 {
		return r.FindByID(id)
	}

	sets = append(sets, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf(
		`UPDATE books SET %s WHERE id = $%d
		 RETURNING id, name, author, year, masterpiece, created_at, updated_at`,
		strings.Join(sets, ", "), idx,
	)

	var b Book
	err := r.db.QueryRowx(query, args...).StructScan(&b)
	if errors.Is(err, sql.ErrNoRows) {
		return Book{}, ErrNotFound
	}
	if err != nil {
		return Book{}, fmt.Errorf("update book: %w", err)
	}
	return b, nil
}

func (r *postgresRepository) Delete(id int64) error {
	res, err := r.db.Exec("DELETE FROM books WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete book: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete book rows: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func safeSortCol(s string) string {
	if allowedSortCols[s] {
		return s
	}
	return "name"
}

func safeSortDir(s string) string {
	if s == "desc" {
		return "DESC"
	}
	return "ASC"
}
