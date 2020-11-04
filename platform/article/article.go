package article

import (
	"errors"
	"go-blog/platform/role"
	"go-blog/platform/user"
	"net/http"
	"time"

	"github.com/go-chi/render"
)

type Article struct {
	ID         int64  `json:"id"`
	User_ID    int64  `json:"-"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	Favorites  int64  `json:"favorites"`
	Created_At int64  `json:"created_at"`
	Updated_At int64  `json:"updated_at"`
}

type ArticlePayload struct {
	*Article
	User *user.UserPayload `json:"user,omitempty"`
}

func NewArticlePayload(article *Article, userRepo *user.Repo, roleRepo *role.Repo) *ArticlePayload {
	payload := &ArticlePayload{Article: article}
	if payload.User == nil && userRepo != nil {
		if userTemp, err := userRepo.GetByID(article.User_ID); err == nil {
			payload.User = user.NewUserPayload(userTemp, roleRepo)
		}
	}
	return payload
}

func NewArticlesOnlyPayload(articles []*Article) []render.Renderer {
	list := []render.Renderer{}
	for _, article := range articles {
		list = append(list, &ArticlePayload{Article: article})
	}
	return list
}

func NewArticleListPayload(articles []*Article, userRepo *user.Repo, roleRepo *role.Repo) []render.Renderer {
	list := []render.Renderer{}
	for _, article := range articles {
		list = append(list, NewArticlePayload(article, userRepo, roleRepo))
	}
	return list
}

func (a *ArticlePayload) Bind(r *http.Request) error {
	//do stuff on payload after 'receive and decode' but before binding data
	if a.Article == nil {
		return errors.New("missing required Article fields.")
	}

	now := time.Now().Unix()
	a.Updated_At = now
	a.Created_At = now

	return nil
}

func (a *ArticlePayload) Render(w http.ResponseWriter, r *http.Request) error {
	//do stuff on payload before send
	return nil
}
