// Books API
//
//	@title			Books API
//	@version		1.0
//	@description	REST API for managing books.
//	@host			localhost:8080
//	@BasePath		/
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/swaggo/swag"

	_ "github.com/lucaslevingston/sipug-go-test/docs"
	"github.com/lucaslevingston/sipug-go-test/internal/book"
	"github.com/lucaslevingston/sipug-go-test/internal/platform/database"
	"github.com/lucaslevingston/sipug-go-test/internal/platform/server"
)

func main() {
	_ = godotenv.Load()

	dbCfg := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "books"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	db := mustConnect(dbCfg)
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("migrations applied")

	repo := book.NewRepository(db)
	svc := book.NewService(repo)
	h := book.NewHandler(svc)

	srv := server.New(server.Config{Port: getEnv("PORT", "8080")})
	r := srv.Router()

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Get("/swagger/doc.json", func(w http.ResponseWriter, _ *http.Request) {
		doc, err := swag.ReadDoc()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(doc))
	})

	r.Get("/docs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(scalarHTML))
	})

	h.RegisterRoutes(r)

	log.Printf("listening on :%s — docs at /docs", getEnv("PORT", "8080"))
	if err := srv.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

const scalarHTML = `<!doctype html>
<html>
  <head>
    <title>Books API</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <script id="api-reference" data-url="/swagger/doc.json"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>`

func mustConnect(cfg database.Config) *sqlx.DB {
	const maxRetries = 10
	for i := range maxRetries {
		db, err := database.Connect(cfg)
		if err == nil {
			return db
		}
		log.Printf("db connect attempt %d/%d failed: %v", i+1, maxRetries, err)
		time.Sleep(2 * time.Second)
	}
	log.Fatal("could not connect to database after retries")
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
