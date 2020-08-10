package handler

import (
	"go-blog/platform/article"
	"go-blog/platform/comment"
	"go-blog/platform/role"
	"go-blog/platform/status"
	"go-blog/platform/user"
	"net/http"

	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
)

func CommentDelete(w http.ResponseWriter, r *http.Request) {
	commentTemp := r.Context().Value(CommentKey).(*comment.Comment)
	commentRepo := r.Context().Value(CommentRepoKey).(*comment.Repo)
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)

	_, claims, _ := jwtauth.FromContext(r.Context())
	tempRole, err := roleRepo.GetByID(int64(claims["role_id"].(float64)))
	if err != nil {
		render.Render(w, r, status.ErrUnauthorized("Incorrect token."))
		return
	}

	if commentTemp.User_ID != int64(claims["user_id"].(float64)) &&
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

func CommentUpdate(w http.ResponseWriter, r *http.Request) {
	commentTemp := r.Context().Value(CommentKey).(*comment.Comment)
	commentPayload := comment.NewCommentPayload(commentTemp, nil, nil)

	if err := render.Bind(r, commentPayload); err != nil {
		render.Render(w, r, status.ErrInvalidRequest(err))
		return
	}
	commentTemp = commentPayload.Comment

	commentRepo := r.Context().Value(CommentRepoKey).(*comment.Repo)
	userRepo := r.Context().Value(UserRepoKey).(*user.Repo)
	roleRepo := r.Context().Value(RoleRepoKey).(*role.Repo)

	_, claims, _ := jwtauth.FromContext(r.Context())
	tempRole, err := roleRepo.GetByID(int64(claims["role_id"].(float64)))
	if err != nil {
		render.Render(w, r, status.ErrUnauthorized("Incorrect token."))
		return
	}

	if commentTemp.User_ID != int64(claims["user_id"].(float64)) && !tempRole.Check(role.CanManageOtherComments) {
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
	comments := commentRepo.GetAllInArticle(articleTemp.ID)
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

	_, claims, _ := jwtauth.FromContext(r.Context())
	commentTemp.User_ID = int64(claims["user_id"].(float64))
	tempRole, err := roleRepo.GetByID(int64(claims["role_id"].(float64)))
	if err != nil {
		render.Render(w, r, status.ErrUnauthorized("Incorrect token."))
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