package handler

import (
	"context"
	"database/sql"
	"go-blog/platform/article"
	"go-blog/platform/errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

func ProvideDatabase(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			repo := article.NewRepo(db)
			ctx := context.WithValue(r.Context(), RepoKey, repo)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ArticleIDContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		repo := r.Context().Value(RepoKey).(*article.Repo)
		var article *article.Article
		var err error

		if articleID := chi.URLParam(r, "articleID"); articleID != "" {
			article, err = repo.GetByID(articleID)
		} else {
			render.Render(w, r, errors.ErrNotFound)
			return
		}
		if err != nil {
			render.Render(w, r, errors.ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), ArticleKey, article)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
