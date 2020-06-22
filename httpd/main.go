package main

import (
	"fmt"
	"go-blog/httpd/handler"
	"go-blog/platform/article"
	"net/http"

	"database/sql"

	"github.com/go-chi/chi"
	_ "github.com/mattn/go-sqlite3"
)

const port = ":3000"

func main() {
	db, err := sql.Open("sqlite3", "./blog.db")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	repo := article.New(db)

	r := chi.NewRouter()

	r.Get("/articles", handler.ArticleGet(repo))
	r.Post("/articles", handler.ArticlePost(repo))

	fmt.Println("Serving on port " + port)
	http.ListenAndServe(port, r)
}
