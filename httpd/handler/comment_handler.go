package handler

import (
	"go-blog/platform/article"
	"go-blog/platform/comment"
	"go-blog/platform/role"
	"go-blog/platform/status"
	"go-blog/platform/user"
	"net/http"

	"github.com/go-chi/render"
)

func CommentDelete(w http.ResponseWriter, r *http.Request) {
	commentTemp := r.Context().Value(CommentKey).(*comment.Comment)
	commentRepo := r.Context().Value(CommentRepoKey).(*comment.Repo)
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)
	claims := r.Context().Value(ClaimsKey).(user.Claims)

	tempRole, err := roleRepo.GetByID(claims.RoleID)
	if err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	}

	if commentTemp.User_ID != claims.UserID &&
		!tempRole.Check(role.CanManageOtherComments) {
		render.Render(w, r, status.ErrUnauthorized("You are not the owner of this comment."))
		return
	}

	if err := commentRepo.Delete(commentTemp.ID); err != nil {
		render.Render(w, r, status.ErrInvalidRequest(err))
		return
	}

	render.Render(w, r, status.DelSuccess())
}

func CommentGetByID(w http.ResponseWriter, r *http.Request) {
	commentTemp := r.Context().Value(CommentKey).(*comment.Comment)
	var userRepo *user.Repo
	var roleRepo *role.Repo

	if r.FormValue("user") != "0" {
		userRepo = r.Context().Value(UserRepoKey).(*user.Repo)
		roleRepo = r.Context().Value(RoleRepoKey).(*role.Repo)
	}

	render.Status(r, http.StatusOK)
	render.Render(w, r, comment.NewCommentPayload(commentTemp, userRepo, roleRepo))
}

func CommentUpdate(w http.ResponseWriter, r *http.Request) {
	commentTemp := r.Context().Value(CommentKey).(*comment.Comment)

	created_at := commentTemp.Created_At

	commentPayload := comment.NewCommentPayload(commentTemp, nil, nil)

	if err := render.Bind(r, commentPayload); err != nil {
		render.Render(w, r, status.ErrInvalidRequest(err))
		return
	}

	commentPayload.Created_At = created_at // keep the created date same as before.

	commentTemp = commentPayload.Comment

	commentRepo := r.Context().Value(CommentRepoKey).(*comment.Repo)
	userRepo := r.Context().Value(UserRepoKey).(*user.Repo)
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)
	claims := r.Context().Value(ClaimsKey).(user.Claims)

	tempRole, err := roleRepo.GetByID(claims.RoleID)
	if err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	}

	if commentTemp.User_ID != claims.UserID && !tempRole.Check(role.CanManageOtherComments) {
		render.Render(w, r, status.ErrUnauthorized("You are not the owner of this comment."))
		return
	}

	if err := commentRepo.Update(commentTemp); err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	}

	render.Status(r, http.StatusOK)
	render.Render(w, r, comment.NewCommentPayload(commentTemp, userRepo, roleRepo))
}

func CommentsGet(w http.ResponseWriter, r *http.Request) {
	articleTemp := r.Context().Value(ArticleKey).(*article.Article)
	commentRepo := r.Context().Value(CommentRepoKey).(*comment.Repo)
	userRepo := r.Context().Value(UserRepoKey).(*user.Repo)
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)

	page := r.Context().Value(PageKey).(int)
	dates := r.Context().Value(DatesKey).([2]int64)

	search := comment.NewSearch()
	search.QueryDate(dates[0], dates[1])
	search.QueryKeyword(r.FormValue("search"))
	search.QueryArticleID(articleTemp.ID)
	search.Limit(page)
	comments := commentRepo.GetMultiple(search)

	render.RenderList(w, r, comment.NewCommentListPayload(comments, userRepo, roleRepo))
}

func CommentPost(w http.ResponseWriter, r *http.Request) {
	data := &comment.CommentPayload{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, status.ErrInvalidRequest(err))
		return
	}

	commentTemp := data.Comment

	commentRepo := r.Context().Value(CommentRepoKey).(*comment.Repo)
	userRepo := r.Context().Value(UserRepoKey).(*user.Repo)
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)
	claims := r.Context().Value(ClaimsKey).(user.Claims)

	commentTemp.User_ID = claims.UserID
	tempRole, err := roleRepo.GetByID(claims.RoleID)
	if err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	}

	if !tempRole.Check(role.CanComment) {
		render.Render(w, r, status.ErrUnauthorized("You don't have enough authority to comment."))
		return
	}

	articleTemp := r.Context().Value(ArticleKey).(*article.Article)
	commentTemp.Article_ID = articleTemp.ID

	if id, err := commentRepo.Add(commentTemp); err != nil {
		render.Render(w, r, status.ErrInternal(err))
		return
	} else {
		commentTemp.ID = id
	}

	render.Status(r, http.StatusCreated)
	render.Render(w, r, comment.NewCommentPayload(commentTemp, userRepo, roleRepo))
}
