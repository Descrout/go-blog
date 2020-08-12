package comment

import (
	"errors"
	"go-blog/platform/role"
	"go-blog/platform/user"
	"net/http"
	"time"

	"github.com/go-chi/render"
)

type Comment struct {
	ID         int64  `json:"id"`
	User_ID    int64  `json:"-"`
	Article_ID int64  `json:"-"`
	Body       string `json:"body"`
	Created_At int64  `json:"created_at"`
	Updated_At int64  `json:"updated_at"`
}

type CommentPayload struct {
	*Comment
	User *user.UserPayload `json:"user,omitempty"`
}

func NewCommentPayload(comment *Comment, userRepo *user.Repo, roleRepo *role.Repo) *CommentPayload {
	payload := &CommentPayload{Comment: comment}
	if payload.User == nil && userRepo != nil {
		if userTemp, err := userRepo.GetByID(comment.User_ID); err == nil {
			payload.User = user.NewUserPayload(userTemp, roleRepo)
		}
	}
	return payload
}

func NewCommentListPayload(comments []*Comment, userRepo *user.Repo, roleRepo *role.Repo) []render.Renderer {
	list := []render.Renderer{}
	for _, comment := range comments {
		list = append(list, NewCommentPayload(comment, userRepo, roleRepo))
	}
	return list
}

func (c *CommentPayload) Bind(r *http.Request) error {
	//do stuff on payload after 'receive and decode' but before binding data
	if c.Comment == nil {
		return errors.New("missing required Comment fields.")
	}
	now := time.Now().Unix()
	c.Updated_At = now
	c.Created_At = now
	return nil
}

func (c *CommentPayload) Render(w http.ResponseWriter, r *http.Request) error {
	//do stuff on payload before send
	return nil
}
