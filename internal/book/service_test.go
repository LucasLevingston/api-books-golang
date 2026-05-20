package book_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/lucaslevingston/sipug-go-test/internal/book"
)

// repoMock implements book.Repository for unit testing the service.
type repoMock struct{ mock.Mock }

func (m *repoMock) FindAll(f book.ListFilter) (book.PageResult, error) {
	a := m.Called(f)
	return a.Get(0).(book.PageResult), a.Error(1)
}
func (m *repoMock) FindByID(id int64) (book.Book, error) {
	a := m.Called(id)
	return a.Get(0).(book.Book), a.Error(1)
}
func (m *repoMock) Create(i book.CreateInput) (book.Book, error) {
	a := m.Called(i)
	return a.Get(0).(book.Book), a.Error(1)
}
func (m *repoMock) Update(id int64, i book.UpdateInput) (book.Book, error) {
	a := m.Called(id, i)
	return a.Get(0).(book.Book), a.Error(1)
}
func (m *repoMock) Delete(id int64) error {
	return m.Called(id).Error(0)
}

func TestService_List_NormalizeDefaults(t *testing.T) {
	repo := new(repoMock)
	want := book.PageResult{Data: []book.Book{}, Total: 0, Page: 1, PageSize: 10, TotalPages: 0}
	repo.On("FindAll", book.ListFilter{Page: 1, PageSize: 10, SortBy: "name", SortDir: "asc"}).Return(want, nil)

	got, err := book.NewService(repo).List(book.ListFilter{})

	assert.NoError(t, err)
	assert.Equal(t, 1, got.Page)
	assert.Equal(t, 10, got.PageSize)
	repo.AssertExpectations(t)
}

func TestService_List_ClampOversizedPageSize(t *testing.T) {
	repo := new(repoMock)
	want := book.PageResult{Data: []book.Book{}, Total: 0, Page: 1, PageSize: 10, TotalPages: 0}
	repo.On("FindAll", book.ListFilter{Page: 1, PageSize: 10, SortBy: "name", SortDir: "asc"}).Return(want, nil)

	_, err := book.NewService(repo).List(book.ListFilter{PageSize: 9999})

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestService_List_InvalidSortByFallsBackToName(t *testing.T) {
	repo := new(repoMock)
	want := book.PageResult{Data: []book.Book{}, Total: 0, Page: 1, PageSize: 10, TotalPages: 0}
	repo.On("FindAll", book.ListFilter{Page: 1, PageSize: 10, SortBy: "name", SortDir: "asc"}).Return(want, nil)

	_, err := book.NewService(repo).List(book.ListFilter{SortBy: "drop table books; --"})

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestService_List_ValidSortByPreserved(t *testing.T) {
	repo := new(repoMock)
	want := book.PageResult{Data: []book.Book{}, Total: 0, Page: 1, PageSize: 10, TotalPages: 0}
	repo.On("FindAll", book.ListFilter{Page: 1, PageSize: 10, SortBy: "author", SortDir: "desc"}).Return(want, nil)

	_, err := book.NewService(repo).List(book.ListFilter{SortBy: "author", SortDir: "desc"})

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestService_Get_NotFound(t *testing.T) {
	repo := new(repoMock)
	repo.On("FindByID", int64(99)).Return(book.Book{}, book.ErrNotFound)

	_, err := book.NewService(repo).Get(99)

	assert.ErrorIs(t, err, book.ErrNotFound)
	repo.AssertExpectations(t)
}

func TestService_Create_DelegatesToRepo(t *testing.T) {
	repo := new(repoMock)
	input := book.CreateInput{Name: "1984", Author: "George Orwell", Year: 1949}
	expected := book.Book{ID: 1, Name: "1984", Author: "George Orwell", Year: 1949}
	repo.On("Create", input).Return(expected, nil)

	got, err := book.NewService(repo).Create(input)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), got.ID)
	repo.AssertExpectations(t)
}

func TestService_Update_NotFound(t *testing.T) {
	repo := new(repoMock)
	name := "New Title"
	input := book.UpdateInput{Name: &name}
	repo.On("Update", int64(99), input).Return(book.Book{}, book.ErrNotFound)

	_, err := book.NewService(repo).Update(99, input)

	assert.ErrorIs(t, err, book.ErrNotFound)
	repo.AssertExpectations(t)
}

func TestService_Delete_NotFound(t *testing.T) {
	repo := new(repoMock)
	repo.On("Delete", int64(99)).Return(book.ErrNotFound)

	err := book.NewService(repo).Delete(99)

	assert.ErrorIs(t, err, book.ErrNotFound)
	repo.AssertExpectations(t)
}

func TestService_Delete_Success(t *testing.T) {
	repo := new(repoMock)
	repo.On("Delete", int64(1)).Return(nil)

	err := book.NewService(repo).Delete(1)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}
