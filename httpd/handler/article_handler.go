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

func ArticleDelete(w http.ResponseWriter, r *http.Request) {
	articleTemp := r.Context().Value(ArticleKey).(*article.Article)
	repo := r.Context().Value(RepoKey).(*article.Repo)

	if err := repo.Delete(articleTemp.ID); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Status(r, http.StatusOK)
	render.Render(w, r, article.NewArticlePayload(articleTemp))
}

func ArticleUpdate(w http.ResponseWriter, r *http.Request) {
	articleTemp := r.Context().Value(ArticleKey).(*article.Article)
	articlePayload := article.NewArticlePayload(articleTemp)

	if err := render.Bind(r, articlePayload); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	articleTemp = articlePayload.Article
	repo := r.Context().Value(RepoKey).(*article.Repo)

	if err := repo.Update(articleTemp); err != nil {
		render.Render(w, r, ErrInternal(err))
		return
	}

	render.Status(r, http.StatusOK)
	render.Render(w, r, article.NewArticlePayload(articleTemp))
}

func ArticleGetByID(w http.ResponseWriter, r *http.Request) {
	articleTemp := r.Context().Value(ArticleKey).(*article.Article)
	render.Render(w, r, article.NewArticlePayload(articleTemp))
}

func ArticleGetAll(w http.ResponseWriter, r *http.Request) {
	repo := r.Context().Value(RepoKey).(*article.Repo)
	articles := repo.GetAll()
	render.RenderList(w, r, article.NewArticleListPayload(articles))
}

func ArticlePost(w http.ResponseWriter, r *http.Request) {
	data := &article.ArticlePayload{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	articleTemp := data.Article

	repo := r.Context().Value(RepoKey).(*article.Repo)

	if id, err := repo.Add(articleTemp); err != nil {
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
