package handler

import (
	"context"
	"go-blog/platform/article"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type key int

const RepoKey key = 0
const ArticleKey key = 1

func ArticleGetByID(w http.ResponseWriter, r *http.Request) {
	article_temp := r.Context().Value(ArticleKey).(*article.Article)
	render.Render(w, r, article.NewArticlePayload(article_temp))
}

func ArticleGetAll(w http.ResponseWriter, r *http.Request) {
	repo := r.Context().Value(RepoKey).(*article.Repo)
	articles := repo.GetAll()
	//json.NewEncoder(w).Encode(articles)
	render.RenderList(w, r, article.NewArticleListPayload(articles))
}

func ArticlePost(w http.ResponseWriter, r *http.Request) {
	data := &article.ArticlePayload{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	article_temp := data.Article

	repo := r.Context().Value(RepoKey).(*article.Repo)

	if id, err := repo.Add(article_temp); err != nil {
		render.Render(w, r, ErrInternal(err))
		return
	} else {
		data.Article.ID = id
	}

	render.Status(r, http.StatusCreated)
	render.Render(w, r, data)
}

func ArticleIDContext(next http.Handler) http.Handler {
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
