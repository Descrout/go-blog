package article

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
)

type Article struct {
	ID      int    `json:"id"`
	User_ID int    `json:"user_id"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	Date    string `json:"date"`
}

type ArticlePayload struct {
	*Article
	//add user payload later
}

func NewArticlePayload(article *Article) *ArticlePayload {
	return &ArticlePayload{Article: article}
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
	a.Date = "00.00.00" // will change to server time later
	a.User_ID = 0       //will change to actual user
	return nil
}

func (a *ArticlePayload) Render(w http.ResponseWriter, r *http.Request) error {
	//do stuff on payload before send
	return nil
}
