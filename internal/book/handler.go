package book

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	svc      Service
	validate *validator.Validate
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc, validate: validator.New()}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/books", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Put("/{id}", h.update)
		r.Delete("/{id}", h.delete)
	})
}

// list godoc
//
//	@Summary		List books
//	@Description	Returns a paginated list of books with optional filters.
//	@Tags			books
//	@Produce		json
//	@Param			author		query		string	false	"Filter by author (partial, case-insensitive)"
//	@Param			masterpiece	query		bool	false	"Filter masterpieces only"
//	@Param			year		query		int		false	"Filter by year"
//	@Param			page		query		int		false	"Page number (default 1)"
//	@Param			page_size	query		int		false	"Items per page, 1-100 (default 10)"
//	@Param			sort_by		query		string	false	"Sort field"		Enums(name,author,year,masterpiece)
//	@Param			sort_dir	query		string	false	"Sort direction"	Enums(asc,desc)
//	@Success		200			{object}	PageResult
//	@Failure		400			{object}	errorResponse
//	@Failure		500			{object}	errorResponse
//	@Router			/books [get]
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := ListFilter{
		Author:  q.Get("author"),
		SortBy:  q.Get("sort_by"),
		SortDir: q.Get("sort_dir"),
	}
	filter.Page, _ = strconv.Atoi(q.Get("page"))
	filter.PageSize, _ = strconv.Atoi(q.Get("page_size"))

	if v := q.Get("masterpiece"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			jsonError(w, http.StatusBadRequest, "masterpiece must be true or false")
			return
		}
		filter.Masterpiece = &b
	}
	if v := q.Get("year"); v != "" {
		y, err := strconv.Atoi(v)
		if err != nil {
			jsonError(w, http.StatusBadRequest, "year must be an integer")
			return
		}
		filter.Year = &y
	}

	result, err := h.svc.List(filter)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// create godoc
//
//	@Summary		Create a book
//	@Description	Creates a new book and returns the created resource.
//	@Tags			books
//	@Accept			json
//	@Produce		json
//	@Param			body	body		CreateInput	true	"Book payload"
//	@Success		201		{object}	Book
//	@Failure		400		{object}	errorResponse
//	@Failure		422		{object}	validationErrorResponse
//	@Failure		500		{object}	errorResponse
//	@Router			/books [post]
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var input CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(input); err != nil {
		jsonValidationError(w, err)
		return
	}
	b, err := h.svc.Create(input)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusCreated, b)
}

// update godoc
//
//	@Summary		Update a book
//	@Description	Partially updates a book by ID. All fields are optional.
//	@Tags			books
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int			true	"Book ID"
//	@Param			body	body		UpdateInput	true	"Fields to update"
//	@Success		200		{object}	Book
//	@Failure		400		{object}	errorResponse
//	@Failure		404		{object}	errorResponse
//	@Failure		422		{object}	validationErrorResponse
//	@Failure		500		{object}	errorResponse
//	@Router			/books/{id} [put]
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "id must be a positive integer")
		return
	}
	var input UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(input); err != nil {
		jsonValidationError(w, err)
		return
	}
	b, err := h.svc.Update(id, input)
	if errors.Is(err, ErrNotFound) {
		jsonError(w, http.StatusNotFound, "book not found")
		return
	}
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, b)
}

// delete godoc
//
//	@Summary		Delete a book
//	@Description	Deletes a book by ID.
//	@Tags			books
//	@Produce		json
//	@Param			id	path	int	true	"Book ID"
//	@Success		204
//	@Failure		400	{object}	errorResponse
//	@Failure		404	{object}	errorResponse
//	@Failure		500	{object}	errorResponse
//	@Router			/books/{id} [delete]
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "id must be a positive integer")
		return
	}
	if err := h.svc.Delete(id); errors.Is(err, ErrNotFound) {
		jsonError(w, http.StatusNotFound, "book not found")
		return
	} else if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}

func jsonResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// errorResponse is the standard error envelope.
type errorResponse struct {
	Error string `json:"error"`
}

// validationErrorResponse is returned on 422 with per-field details.
type validationErrorResponse struct {
	Errors []fieldError `json:"errors"`
}

type fieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func jsonError(w http.ResponseWriter, status int, msg string) {
	jsonResponse(w, status, errorResponse{Error: msg})
}

func jsonValidationError(w http.ResponseWriter, err error) {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		jsonError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	errs := make([]fieldError, len(ve))
	for i, fe := range ve {
		errs[i] = fieldError{
			Field:   strings.ToLower(fe.Field()),
			Message: validationMessage(fe),
		}
	}
	jsonResponse(w, http.StatusUnprocessableEntity, validationErrorResponse{Errors: errs})
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "min":
		return fmt.Sprintf("%s must be at least %s", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", fe.Field(), fe.Param())
	default:
		return fe.Field() + " is invalid"
	}
}
