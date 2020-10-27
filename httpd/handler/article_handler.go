package handler

import (
	"go-blog/platform/article"
	"go-blog/platform/role"
	"go-blog/platform/status"
	"go-blog/platform/user"
	"net/http"

	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
)

func ArticleDelete(w http.ResponseWriter, r *http.Request) {
	articleTemp := r.Context().Value(ArticleKey).(*article.Article)
	articleRepo := r.Context().Value(ArticleRepoKey).(*article.Repo)
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)

	_, claims, _ := jwtauth.FromContext(r.Context())
	tempRole, err := roleRepo.GetByID(claims["role_id"])
	if err != nil {
		render.Render(w, r, status.ErrUnauthorized("Incorrect token."))
		return
	}

	if articleTemp.User_ID != int64(claims["user_id"].(float64)) && !tempRole.Check(role.CanManageOtherArticle) {
		render.Render(w, r, status.ErrUnauthorized("You are not the author."))
		return
	}

	if err := articleRepo.Delete(articleTemp.ID); err != nil {
		render.Render(w, r, status.ErrInvalidRequest(err))
		return
	}

	render.Render(w, r, status.DelSuccess())
}

func ArticleUpdate(w http.ResponseWriter, r *http.Request) {
	articleTemp := r.Context().Value(ArticleKey).(*article.Article)

	created_at := articleTemp.Created_At

	articlePayload := article.NewArticlePayload(articleTemp, nil, nil)

	if err := render.Bind(r, articlePayload); err != nil {
		render.Render(w, r, status.ErrInvalidRequest(err))
		return
	}

	articlePayload.Created_At = created_at // keep the created date same as before.

	articleTemp = articlePayload.Article

	articleRepo := r.Context().Value(ArticleRepoKey).(*article.Repo)
	userRepo := r.Context().Value(UserRepoKey).(*user.Repo)
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)

	_, claims, _ := jwtauth.FromContext(r.Context())
	tempRole, err := roleRepo.GetByID(claims["role_id"])
	if err != nil {
		render.Render(w, r, status.ErrUnauthorized("Incorrect token."))
		return
	}

	if articleTemp.User_ID != int64(claims["user_id"].(float64)) && !tempRole.Check(role.CanManageOtherArticle) {
		render.Render(w, r, status.ErrUnauthorized("You are not the author."))
		return
	}

	if err := articleRepo.Update(articleTemp); err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	}

	render.Status(r, http.StatusOK)
	render.Render(w, r, article.NewArticlePayload(articleTemp, userRepo, roleRepo))
}

func ArticleGetByID(w http.ResponseWriter, r *http.Request) {
	articleTemp := r.Context().Value(ArticleKey).(*article.Article)
	userRepo := r.Context().Value(UserRepoKey).(*user.Repo)
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)
	render.Render(w, r, article.NewArticlePayload(articleTemp, userRepo, roleRepo))
}

func ArticleGetMultiple(w http.ResponseWriter, r *http.Request) {
	articleRepo := r.Context().Value(ArticleRepoKey).(*article.Repo)
	userRepo := r.Context().Value(UserRepoKey).(*user.Repo)
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)
	page := r.Context().Value(PageKey).(int)
	dates := r.Context().Value(DatesKey).([2]int64)

	articles := articleRepo.GetMultiple(page, r.FormValue("search"), dates)

	render.RenderList(w, r, article.NewArticleListPayload(articles, userRepo, roleRepo))
}

func ArticlePost(w http.ResponseWriter, r *http.Request) {
	data := &article.ArticlePayload{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, status.ErrInvalidRequest(err))
		return
	}

	articleTemp := data.Article

	articleRepo := r.Context().Value(ArticleRepoKey).(*article.Repo)
	userRepo := r.Context().Value(UserRepoKey).(*user.Repo)
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)

	_, claims, _ := jwtauth.FromContext(r.Context())
	articleTemp.User_ID = int64(claims["user_id"].(float64))
	tempRole, err := roleRepo.GetByID(claims["role_id"])
	if err != nil {
		render.Render(w, r, status.ErrUnauthorized("Incorrect token."))
		return
	}

	if !tempRole.Check(role.CanPostArticle) {
		render.Render(w, r, status.ErrUnauthorized("You don't have enough authority to post an article."))
		return
	}

	if id, err := articleRepo.Add(articleTemp); err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	} else {
		articleTemp.ID = id
	}

	render.Status(r, http.StatusCreated)
	render.Render(w, r, article.NewArticlePayload(data.Article, userRepo, roleRepo))
}
