package handler

import (
	"go-blog/platform/article"
	"go-blog/platform/errors"
	"net/http"

	"github.com/go-chi/render"
)

func ArticleDelete(w http.ResponseWriter, r *http.Request) {
	articleTemp := r.Context().Value(ArticleKey).(*article.Article)
	repo := r.Context().Value(ArticleRepoKey).(*article.Repo)

	if err := repo.Delete(articleTemp.ID); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	render.Status(r, http.StatusOK)
	render.Render(w, r, article.NewArticlePayload(articleTemp))
}

func ArticleUpdate(w http.ResponseWriter, r *http.Request) {
	articleTemp := r.Context().Value(ArticleKey).(*article.Article)
	articlePayload := article.NewArticlePayload(articleTemp)

	if err := render.Bind(r, articlePayload); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	articleTemp = articlePayload.Article
	repo := r.Context().Value(ArticleRepoKey).(*article.Repo)

	if err := repo.Update(articleTemp); err != nil {
		render.Render(w, r, errors.ErrInternal(err))
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
	repo := r.Context().Value(ArticleRepoKey).(*article.Repo)
	articles := repo.GetAll()
	render.RenderList(w, r, article.NewArticleListPayload(articles))
}

func ArticlePost(w http.ResponseWriter, r *http.Request) {
	data := &article.ArticlePayload{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, errors.ErrInvalidRequest(err))
		return
	}

	articleTemp := data.Article

	repo := r.Context().Value(ArticleRepoKey).(*article.Repo)

	if id, err := repo.Add(articleTemp); err != nil {
		render.Render(w, r, errors.ErrInternal(err))
		return
	} else {
		data.Article.ID = id
	}

	render.Status(r, http.StatusCreated)
	render.Render(w, r, data)
}
