package handler

import (
	"context"
	"encoding/json"
	"go-blog/platform/article"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type key int

const RepoKey key = 0
const ArticleKey key = 1

func ArticleGetByID(w http.ResponseWriter, r *http.Request) {
	article := r.Context().Value(ArticleKey).(*article.Article)
	json.NewEncoder(w).Encode(article)
}

func ArticleGetAll(w http.ResponseWriter, r *http.Request) {
	repo := r.Context().Value(RepoKey).(*article.Repo)
	articles := repo.GetAll()
	json.NewEncoder(w).Encode(articles)
}

func ArticlePost(w http.ResponseWriter, r *http.Request) {
	repo := r.Context().Value(RepoKey).(*article.Repo)
	request := map[string]string{}
	json.NewDecoder(r.Body).Decode(&request)
	repo.Add(article.Article{
		User_ID: 0, //change later
		Title:   request["title"],
		Body:    request["body"],
		Date:    "00.00.00", //change later
	})
	w.Write([]byte("Article added!"))
}

func ArticleIDCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		repo := r.Context().Value(RepoKey).(*article.Repo)
		var article *article.Article
		var err error

		if articleID := chi.URLParam(r, "articleID"); articleID != "" {
			article, err = repo.GetByID(articleID)
		} else {
			render.Render(w, r, ErrNotFound)
			return
		}
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), ArticleKey, article)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
