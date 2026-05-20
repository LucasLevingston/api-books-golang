//go:build integration

package book_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/lucaslevingston/api-books-golang/internal/book"
	"github.com/lucaslevingston/api-books-golang/internal/platform/database"
)

var testDB *sqlx.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("books_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("start postgres container: %v", err)
	}
	defer pgContainer.Terminate(ctx) //nolint:errcheck

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("get connection string: %v", err)
	}

	testDB, err = sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatalf("connect to postgres: %v", err)
	}
	defer testDB.Close()

	if err := database.Migrate(testDB); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	os.Exit(m.Run())
}

func setupRepo(t *testing.T) book.Repository {
	t.Helper()
	_, err := testDB.Exec("TRUNCATE TABLE books RESTART IDENTITY CASCADE")
	require.NoError(t, err)
	return book.NewRepository(testDB)
}

func TestRepository_Create(t *testing.T) {
	repo := setupRepo(t)

	b, err := repo.Create(book.CreateInput{Name: "1984", Author: "George Orwell", Year: 1949, Masterpiece: true})

	require.NoError(t, err)
	assert.NotZero(t, b.ID)
	assert.Equal(t, "1984", b.Name)
	assert.Equal(t, "George Orwell", b.Author)
	assert.Equal(t, 1949, b.Year)
	assert.True(t, b.Masterpiece)
	assert.NotZero(t, b.CreatedAt)
}

func TestRepository_FindByID_Found(t *testing.T) {
	repo := setupRepo(t)
	created, err := repo.Create(book.CreateInput{Name: "Dune", Author: "Frank Herbert", Year: 1965})
	require.NoError(t, err)

	found, err := repo.FindByID(created.ID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "Dune", found.Name)
}

func TestRepository_FindByID_NotFound(t *testing.T) {
	repo := setupRepo(t)

	_, err := repo.FindByID(999999)

	assert.ErrorIs(t, err, book.ErrNotFound)
}

func TestRepository_FindAll_Empty(t *testing.T) {
	repo := setupRepo(t)

	result, err := repo.FindAll(book.ListFilter{Page: 1, PageSize: 10, SortBy: "name", SortDir: "asc"})

	require.NoError(t, err)
	assert.Equal(t, 0, result.Total)
	assert.Empty(t, result.Data)
}

func TestRepository_FindAll_FilterByAuthor(t *testing.T) {
	repo := setupRepo(t)
	repo.Create(book.CreateInput{Name: "1984", Author: "George Orwell", Year: 1949})       //nolint:errcheck
	repo.Create(book.CreateInput{Name: "The Hobbit", Author: "J.R.R. Tolkien", Year: 1937}) //nolint:errcheck

	result, err := repo.FindAll(book.ListFilter{Author: "orwell", Page: 1, PageSize: 10, SortBy: "name", SortDir: "asc"})

	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, "1984", result.Data[0].Name)
}

func TestRepository_FindAll_FilterByMasterpiece(t *testing.T) {
	repo := setupRepo(t)
	yes := true
	repo.Create(book.CreateInput{Name: "Masterpiece", Author: "A", Year: 2000, Masterpiece: true}) //nolint:errcheck
	repo.Create(book.CreateInput{Name: "Regular", Author: "B", Year: 2001})                        //nolint:errcheck

	result, err := repo.FindAll(book.ListFilter{Masterpiece: &yes, Page: 1, PageSize: 10, SortBy: "name", SortDir: "asc"})

	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, "Masterpiece", result.Data[0].Name)
}

func TestRepository_FindAll_FilterByYear(t *testing.T) {
	repo := setupRepo(t)
	year := 1949
	repo.Create(book.CreateInput{Name: "1984", Author: "Orwell", Year: 1949})  //nolint:errcheck
	repo.Create(book.CreateInput{Name: "Dune", Author: "Herbert", Year: 1965}) //nolint:errcheck

	result, err := repo.FindAll(book.ListFilter{Year: &year, Page: 1, PageSize: 10, SortBy: "name", SortDir: "asc"})

	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, "1984", result.Data[0].Name)
}

func TestRepository_FindAll_Pagination(t *testing.T) {
	repo := setupRepo(t)
	for i := range 5 {
		repo.Create(book.CreateInput{Name: fmt.Sprintf("Book %02d", i), Author: "A", Year: 2000}) //nolint:errcheck
	}

	result, err := repo.FindAll(book.ListFilter{Page: 2, PageSize: 2, SortBy: "name", SortDir: "asc"})

	require.NoError(t, err)
	assert.Equal(t, 5, result.Total)
	assert.Len(t, result.Data, 2)
	assert.Equal(t, 3, result.TotalPages)
	assert.Equal(t, "Book 02", result.Data[0].Name)
}

func TestRepository_FindAll_SortDesc(t *testing.T) {
	repo := setupRepo(t)
	repo.Create(book.CreateInput{Name: "A Book", Author: "A", Year: 2000}) //nolint:errcheck
	repo.Create(book.CreateInput{Name: "Z Book", Author: "Z", Year: 2001}) //nolint:errcheck

	result, err := repo.FindAll(book.ListFilter{Page: 1, PageSize: 10, SortBy: "name", SortDir: "desc"})

	require.NoError(t, err)
	assert.Equal(t, "Z Book", result.Data[0].Name)
}

func TestRepository_Update_Success(t *testing.T) {
	repo := setupRepo(t)
	created, err := repo.Create(book.CreateInput{Name: "Original", Author: "Author", Year: 2000})
	require.NoError(t, err)

	newName := "Updated"
	updated, err := repo.Update(created.ID, book.UpdateInput{Name: &newName})

	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Name)
	assert.Equal(t, created.Author, updated.Author)
}

func TestRepository_Update_AllFields(t *testing.T) {
	repo := setupRepo(t)
	created, err := repo.Create(book.CreateInput{Name: "Old", Author: "Old Author", Year: 2000})
	require.NoError(t, err)

	newName := "New"
	newAuthor := "New Author"
	newYear := 2024
	yes := true
	updated, err := repo.Update(created.ID, book.UpdateInput{
		Name:        &newName,
		Author:      &newAuthor,
		Year:        &newYear,
		Masterpiece: &yes,
	})

	require.NoError(t, err)
	assert.Equal(t, "New", updated.Name)
	assert.Equal(t, "New Author", updated.Author)
	assert.Equal(t, 2024, updated.Year)
	assert.True(t, updated.Masterpiece)
}

func TestRepository_Update_NoFields_ReturnsCurrent(t *testing.T) {
	repo := setupRepo(t)
	created, err := repo.Create(book.CreateInput{Name: "Same", Author: "Author", Year: 2000})
	require.NoError(t, err)

	found, err := repo.Update(created.ID, book.UpdateInput{})

	require.NoError(t, err)
	assert.Equal(t, "Same", found.Name)
}

func TestRepository_Update_NotFound(t *testing.T) {
	repo := setupRepo(t)
	name := "New"

	_, err := repo.Update(999999, book.UpdateInput{Name: &name})

	assert.ErrorIs(t, err, book.ErrNotFound)
}

func TestRepository_Delete_Success(t *testing.T) {
	repo := setupRepo(t)
	created, err := repo.Create(book.CreateInput{Name: "To Delete", Author: "Author", Year: 2000})
	require.NoError(t, err)

	err = repo.Delete(created.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(created.ID)
	assert.ErrorIs(t, err, book.ErrNotFound)
}

func TestRepository_Delete_NotFound(t *testing.T) {
	repo := setupRepo(t)

	err := repo.Delete(999999)

	assert.ErrorIs(t, err, book.ErrNotFound)
}
