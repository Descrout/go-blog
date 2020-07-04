package article

import (
	"errors"
	"go-blog/platform/user"
	"net/http"
	"time"

	"github.com/go-chi/render"
)

type Article struct {
	ID      int64  `json:"id"`
	User_ID int64  `json:"user_id,omitempty"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	Date    int64  `json:"date"`
}

type ArticlePayload struct {
	*Article
	User *user.UserPayload `json:"user"`
}

func NewArticlePayload(article *Article) *ArticlePayload {
	payload := &ArticlePayload{Article: article}
	//TODO User payloading
	return payload
}

func NewArticleListPayload(articles []*Article) []render.Renderer {
	list := []render.Renderer{}
	for _, article := range articles {
		list = append(list, NewArticlePayload(article))
	}
	return list
}

func (a *ArticlePayload) Bind(r *http.Request) error {
	//do stuff on payload after 'receive and decode' but before binding data
	if a.Article == nil {
		return errors.New("missing required Article fields.")
	}
	a.Date = time.Now().Unix()
	return nil
}

func (a *ArticlePayload) Render(w http.ResponseWriter, r *http.Request) error {
	//do stuff on payload before send

	a.User_ID = 0 // we set id to 0 so we won't send the id (we already include the user)
	return nil
}
