package main

import (
	"context"
	"database/sql"
	"fmt"
	"go-blog/httpd/handler"
	"go-blog/platform/article"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/mattn/go-sqlite3"
)

const (
	port   = ":3000"
	dbName = "./blog.db"
)

func main() {
	db := setupDB(dbName)
	defer db.Close()

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("TODO"))
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("You page is in another castle."))
	})

	r.Route("/articles", func(r chi.Router) {
		r.Use(ProvideArticlesRepo(db))
		r.Get("/", handler.ArticleGetAll)
		//r.Get("/{month}-{day}-{year}", listArticlesByDate)

		r.Post("/", handler.ArticlePost)
		//r.Get("/search", searchArticles)

		// Subrouters:
		r.Route("/{articleID}", func(r chi.Router) {
			r.Use(handler.ArticleIDContext)
			r.Get("/", handler.ArticleGetByID)
			//r.Put("/", updateArticle)
			//r.Delete("/", deleteArticle)
		})
	})

	fmt.Println("Serving on port " + port)
	http.ListenAndServe(port, r)
}

func ProvideArticlesRepo(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			repo := article.NewRepo(db)
			ctx := context.WithValue(r.Context(), handler.RepoKey, repo)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func setupDB(filename string) *sql.DB {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := db.Prepare(`CREATE TABLE IF NOT EXISTS "articles" (
		"id"	INTEGER NOT NULL UNIQUE,
		"user_id"	INTEGER NOT NULL,
		"title"	TEXT NOT NULL,
		"body"	TEXT NOT NULL,
		"date"	TEXT NOT NULL,
		PRIMARY KEY("ID" AUTOINCREMENT)
	);`)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	stmt.Exec()

	return db
}
