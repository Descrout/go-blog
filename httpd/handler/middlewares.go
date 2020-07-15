package handler

import (
	"context"
	"database/sql"
	"errors"
	"go-blog/platform/article"
	"go-blog/platform/comment"
	"go-blog/platform/role"
	"go-blog/platform/status"
	"go-blog/platform/user"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type key int

const (
	ArticleRepoKey key = 0
	ArticleKey     key = 1
	UserRepoKey    key = 2
	UserKey        key = 3
	RoleRepoKey    key = 4
	CommentRepoKey key = 5
	CommentKey     key = 6
)

func ProvideCommentRepo(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			repo := comment.NewRepo(db)
			ctx := context.WithValue(r.Context(), CommentRepoKey, repo)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ProvideArticleRepo(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			repo := article.NewRepo(db)
			ctx := context.WithValue(r.Context(), ArticleRepoKey, repo)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ProvideRoleRepo(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			repo := role.NewRepo(db)
			ctx := context.WithValue(r.Context(), RoleRepoKey, repo)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ProvideUserRepo(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			repo := user.NewRepo(db)
			ctx := context.WithValue(r.Context(), UserRepoKey, repo)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CommentIDContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		repo := r.Context().Value(CommentRepoKey).(*comment.Repo)
		var comment *comment.Comment
		var err error

		if commentID := chi.URLParam(r, "commentID"); commentID != "" {
			comment, err = repo.GetByID(commentID)
		} else {
			render.Render(w, r, status.ErrInvalidRequest(errors.New("missing comment ID")))
			return
		}
		if err != nil {
			render.Render(w, r, status.ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), CommentKey, comment)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ArticleIDContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		repo := r.Context().Value(ArticleRepoKey).(*article.Repo)
		var article *article.Article
		var err error

		if articleID := chi.URLParam(r, "articleID"); articleID != "" {
			article, err = repo.GetByID(articleID)
		} else {
			render.Render(w, r, status.ErrInvalidRequest(errors.New("missing article ID")))
			return
		}
		if err != nil {
			render.Render(w, r, status.ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), ArticleKey, article)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
