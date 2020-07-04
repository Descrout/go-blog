package comment

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/render"
)

type Comment struct {
	ID         int64  `json:"id"`
	User_ID    int64  `json:"user_id,omitempty"`
	Article_ID int64  `json:"article_id"`
	Body       string `json:"body"`
	Date       int64  `json:"date"`
}

type CommentPayload struct {
	*Comment
	//add user payload later
}

func NewCommentPayload(comment *Comment) *CommentPayload {
	return &CommentPayload{Comment: comment}
}

func NewCommentListPayload(comments []*Comment) []render.Renderer {
	list := []render.Renderer{}
	for _, comment := range comments {
		list = append(list, NewCommentPayload(comment))
	}
	return list
}

func (c *CommentPayload) Bind(r *http.Request) error {
	//do stuff on payload after 'receive and decode' but before binding data
	if c.Comment == nil {
		return errors.New("missing required Comment fields.")
	}
	c.Date = time.Now().Unix()
	return nil
}

func (c *CommentPayload) Render(w http.ResponseWriter, r *http.Request) error {
	//do stuff on payload before send

	c.User_ID = 0 // we set id to 0 so we won't send the id (we already include the user)
	return nil
}
