package main

import (
	"database/sql"
	"fmt"
	"go-blog/httpd/handler"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
	_ "github.com/mattn/go-sqlite3"
)

const (
	port   = ":3000"
	dbName = "./blog.db"
)

func main() {
	db := setupDB(dbName)
	defer db.Close()

	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

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
		w.Write([]byte("Your page is in another castle."))
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(handler.ProvideDatabase(db))
		r.Use(jwtauth.Verifier(tokenAuth))
		// _, tokenString, _ := tokenAuth.Encode(jwt.MapClaims{"user_id": 123}) //creating token
		//_, claims, _ := jwtauth.FromContext(r.Context()) // getting the token - claims["user_id"]
		r.Route("/articles", func(r chi.Router) {
			r.Get("/", handler.ArticleGetAll)
			//r.Get("/{month}-{day}-{year}", listArticlesByDate)
			//r.Get("/search", searchArticles)
			r.With(jwtauth.Authenticator).Post("/", handler.ArticlePost)

			r.Route("/{articleID}", func(r chi.Router) {
				r.Use(handler.ArticleIDContext)
				r.Get("/", handler.ArticleGetByID)
				r.With(jwtauth.Authenticator).Put("/", handler.ArticleUpdate)
				r.With(jwtauth.Authenticator).Delete("/", handler.ArticleDelete)
			})
		})
	})

	fmt.Println("Serving on port " + port)
	http.ListenAndServe(port, r)
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
