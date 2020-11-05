package handler

import (
	"errors"
	"go-blog/platform/article"
	"go-blog/platform/role"
	"go-blog/platform/status"
	"go-blog/platform/user"
	"net/http"

	"github.com/go-chi/render"
)

func ArticleDelete(w http.ResponseWriter, r *http.Request) {
	articleTemp := r.Context().Value(ArticleKey).(*article.Article)
	articleRepo := r.Context().Value(ArticleRepoKey).(*article.Repo)
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)

	claims := r.Context().Value(ClaimsKey).(user.Claims)
	tempRole, err := roleRepo.GetByID(claims.RoleID)
	if err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	}

	if articleTemp.User_ID != claims.UserID && !tempRole.Check(role.CanManageOtherArticle) {
		render.Render(w, r, status.ErrUnauthorized("You are not the author."))
		return
	}

	if err := articleRepo.Delete(articleTemp.ID); err != nil {
		render.Render(w, r, status.ErrInvalidRequest(err))
		return
	}

	render.Render(w, r, status.DelSuccess())
}

func ArticleToggleFavorite(w http.ResponseWriter, r *http.Request) {
	articleTemp := r.Context().Value(ArticleKey).(*article.Article)
	articleRepo := r.Context().Value(ArticleRepoKey).(*article.Repo)
	claims := r.Context().Value(ClaimsKey).(user.Claims)

	articleID := articleTemp.ID

	if articleTemp.User_ID == claims.UserID {
		render.Render(w, r, status.ErrInvalidRequest(errors.New("You cant like your own article.")))
		return
	}

	var favStatus bool
	var err error
	if favStatus, err = articleRepo.ToggleFavoriteFor(articleID, claims.UserID); err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]bool{"fav_status": favStatus})
}

func ArticleUpdate(w http.ResponseWriter, r *http.Request) {
	articleTemp := r.Context().Value(ArticleKey).(*article.Article)

	created_at := articleTemp.Created_At

	articlePayload := article.NewArticlePayload(articleTemp, user.NotAuthenticated, nil, nil)

	if err := render.Bind(r, articlePayload); err != nil {
		render.Render(w, r, status.ErrInvalidRequest(err))
		return
	}

	articlePayload.Created_At = created_at // keep the created date same as before.

	articleTemp = articlePayload.Article

	articleRepo := r.Context().Value(ArticleRepoKey).(*article.Repo)
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)
	claims := r.Context().Value(ClaimsKey).(user.Claims)

	tempRole, err := roleRepo.GetByID(claims.RoleID)
	if err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	}

	if articleTemp.User_ID != claims.UserID && !tempRole.Check(role.CanManageOtherArticle) {
		render.Render(w, r, status.ErrUnauthorized("You are not the author."))
		return
	}

	if err := articleRepo.Update(articleTemp); err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	}

	render.Status(r, http.StatusOK)
	render.Render(w, r, article.NewArticlePayload(articleTemp, user.NotAuthenticated, nil, nil))
}

func ArticleGetByID(w http.ResponseWriter, r *http.Request) {
	articleTemp := r.Context().Value(ArticleKey).(*article.Article)
	claims := r.Context().Value(ClaimsKey).(user.Claims)
	var userRepo *user.Repo
	var roleRepo *role.Repo
	userRepo = r.Context().Value(UserRepoKey).(*user.Repo)

	if r.FormValue("user") != "0" {
		roleRepo = r.Context().Value(RoleRepoKey).(*role.Repo)
	}

	render.Render(w, r, article.NewArticlePayload(articleTemp, claims, userRepo, roleRepo))
}

func ArticleGetMultiple(w http.ResponseWriter, r *http.Request) {
	articleRepo := r.Context().Value(ArticleRepoKey).(*article.Repo)
	page := r.Context().Value(PageKey).(int)
	dates := r.Context().Value(DatesKey).([2]int64)

	search := article.NewSearch()
	search.QueryDate(dates[0], dates[1])
	search.QueryKeyword(r.FormValue("search"))
	search.Limit(page, r.FormValue("sort"))
	articles := articleRepo.GetMultiple(search)

	claims := r.Context().Value(ClaimsKey).(user.Claims)
	userRepo := r.Context().Value(UserRepoKey).(*user.Repo)

	if r.FormValue("user") == "0" {
		render.RenderList(w, r, article.NewArticleListPayload(articles, claims, userRepo, nil))
		return
	}

	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)
	render.RenderList(w, r, article.NewArticleListPayload(articles, claims, userRepo, roleRepo))
}

func ArticlePost(w http.ResponseWriter, r *http.Request) {
	data := &article.ArticlePayload{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, status.ErrInvalidRequest(err))
		return
	}

	articleTemp := data.Article

	articleRepo := r.Context().Value(ArticleRepoKey).(*article.Repo)
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)
	claims := r.Context().Value(ClaimsKey).(user.Claims)

	articleTemp.User_ID = claims.UserID
	tempRole, err := roleRepo.GetByID(claims.RoleID)
	if err != nil {
		render.Render(w, r, status.ErrInternal(err))
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
	render.Render(w, r, article.NewArticlePayload(data.Article, user.NotAuthenticated, nil, nil))
}
