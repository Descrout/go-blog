package handler

import (
	"encoding/json"
	"go-blog/platform/article"
	"net/http"
)

func ArticleGet(repo article.Getter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		articles := repo.GetAll()
		json.NewEncoder(w).Encode(articles)
	}
}
