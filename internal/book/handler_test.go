package book_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/lucaslevingston/sipug-go-test/internal/book"
)

// svcMock implements book.Service for unit testing the HTTP handler.
type svcMock struct{ mock.Mock }

func (m *svcMock) List(f book.ListFilter) (book.PageResult, error) {
	a := m.Called(f)
	return a.Get(0).(book.PageResult), a.Error(1)
}
func (m *svcMock) Get(id int64) (book.Book, error) {
	a := m.Called(id)
	return a.Get(0).(book.Book), a.Error(1)
}
func (m *svcMock) Create(i book.CreateInput) (book.Book, error) {
	a := m.Called(i)
	return a.Get(0).(book.Book), a.Error(1)
}
func (m *svcMock) Update(id int64, i book.UpdateInput) (book.Book, error) {
	a := m.Called(id, i)
	return a.Get(0).(book.Book), a.Error(1)
}
func (m *svcMock) Delete(id int64) error {
	return m.Called(id).Error(0)
}

func newRouter(svc book.Service) http.Handler {
	r := chi.NewRouter()
	book.NewHandler(svc).RegisterRoutes(r)
	return r
}

func TestHandler_List_ReturnsPage(t *testing.T) {
	svc := new(svcMock)
	result := book.PageResult{
		Data:       []book.Book{{ID: 1, Name: "1984", Author: "George Orwell", Year: 1949}},
		Total:      1,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
	}
	svc.On("List", mock.AnythingOfType("book.ListFilter")).Return(result, nil)

	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	w := httptest.NewRecorder()
	newRouter(svc).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body book.PageResult
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 1, body.Total)
	svc.AssertExpectations(t)
}

func TestHandler_List_InvalidMasterpiece(t *testing.T) {
	svc := new(svcMock)
	req := httptest.NewRequest(http.MethodGet, "/books?masterpiece=notbool", nil)
	w := httptest.NewRecorder()
	newRouter(svc).ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_List_InvalidYear(t *testing.T) {
	svc := new(svcMock)
	req := httptest.NewRequest(http.MethodGet, "/books?year=abc", nil)
	w := httptest.NewRecorder()
	newRouter(svc).ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_Create_Valid(t *testing.T) {
	svc := new(svcMock)
	input := book.CreateInput{Name: "1984", Author: "George Orwell", Year: 1949}
	created := book.Book{ID: 1, Name: "1984", Author: "George Orwell", Year: 1949}
	svc.On("Create", input).Return(created, nil)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newRouter(svc).ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestHandler_Create_InvalidBody(t *testing.T) {
	svc := new(svcMock)
	req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader([]byte("{bad json")))
	w := httptest.NewRecorder()
	newRouter(svc).ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_Create_MissingRequiredField(t *testing.T) {
	svc := new(svcMock)
	payload := map[string]any{"author": "X", "year": 2000}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newRouter(svc).ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestHandler_Update_NotFound(t *testing.T) {
	svc := new(svcMock)
	name := "New Title"
	input := book.UpdateInput{Name: &name}
	svc.On("Update", int64(99), input).Return(book.Book{}, book.ErrNotFound)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/books/99", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newRouter(svc).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestHandler_Update_InvalidID(t *testing.T) {
	svc := new(svcMock)
	req := httptest.NewRequest(http.MethodPut, "/books/abc", bytes.NewReader([]byte("{}")))
	w := httptest.NewRecorder()
	newRouter(svc).ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_Delete_Success(t *testing.T) {
	svc := new(svcMock)
	svc.On("Delete", int64(1)).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/books/1", nil)
	w := httptest.NewRecorder()
	newRouter(svc).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	svc.AssertExpectations(t)
}

func TestHandler_Delete_NotFound(t *testing.T) {
	svc := new(svcMock)
	svc.On("Delete", int64(99)).Return(book.ErrNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/books/99", nil)
	w := httptest.NewRecorder()
	newRouter(svc).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestHandler_Delete_InvalidID(t *testing.T) {
	svc := new(svcMock)
	req := httptest.NewRequest(http.MethodDelete, "/books/abc", nil)
	w := httptest.NewRecorder()
	newRouter(svc).ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_Update_Success(t *testing.T) {
	svc := new(svcMock)
	name := "Animal Farm"
	input := book.UpdateInput{Name: &name}
	updated := book.Book{ID: 1, Name: "Animal Farm", Author: "George Orwell", Year: 1945}
	svc.On("Update", int64(1), input).Return(updated, nil)

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/books/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newRouter(svc).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestHandler_Update_InvalidBody(t *testing.T) {
	svc := new(svcMock)
	req := httptest.NewRequest(http.MethodPut, "/books/1", bytes.NewReader([]byte("{bad json")))
	w := httptest.NewRecorder()
	newRouter(svc).ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_Update_ValidationError(t *testing.T) {
	svc := new(svcMock)
	year := 9999
	payload := book.UpdateInput{Year: &year}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, "/books/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newRouter(svc).ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestHandler_Update_InternalError(t *testing.T) {
	svc := new(svcMock)
	name := "New Title"
	input := book.UpdateInput{Name: &name}
	svc.On("Update", int64(1), input).Return(book.Book{}, errors.New("db error"))

	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/books/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newRouter(svc).ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}
