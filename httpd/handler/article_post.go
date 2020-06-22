package handler

import (
	"encoding/json"
	"go-blog/platform/article"
	"net/http"
)

func ArticlePost(repo article.Adder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request := map[string]string{}
		json.NewDecoder(r.Body).Decode(&request)
		repo.Add(article.Article{
			USER_ID: 0,      //change later
			Author:  "todo", //change later
			Title:   request["title"],
			Body:    request["body"],
			Date:    "00.00.00", //change later
		})

		w.Write([]byte("Article added!"))
	}
}
