package article

import (
	"database/sql"
	"fmt"
)

type Getter interface {
	GetAll() []Article
}

type Adder interface {
	Add(article Article)
}

type Article struct {
	ID      int    `json: "id"`
	USER_ID int    `json: "user_id"`
	Author  string `json: "author"`
	Title   string `json: "title"`
	Body    string `json: "body"`
	Date    string `json: "date"`
}

type Repo struct {
	DB *sql.DB
}

func New(db *sql.DB) *Repo {
	stmt, err := db.Prepare(`CREATE TABLE IF NOT EXISTS "articles" (
		"ID"	INTEGER NOT NULL UNIQUE,
		"USER_ID"	INTEGER NOT NULL,
		"author"	TEXT NOT NULL,
		"title"	TEXT NOT NULL,
		"body"	TEXT NOT NULL,
		"date"	TEXT NOT NULL,
		PRIMARY KEY("ID" AUTOINCREMENT)
	);`)
	if err != nil {
		panic(err.Error())
	}
	stmt.Exec()

	return &Repo{
		DB: db,
	}
}

func (repo *Repo) Add(article Article) {
	stmt, err := repo.DB.Prepare(`
	INSERT INTO 
	articles (USER_ID, author, title, body, date) 
	values (?, ?, ?, ?, ?)`)

	if err != nil {
		fmt.Println(err)
	}
	stmt.Exec(article.USER_ID, article.Author, article.Title, article.Body, article.Date)
}

func (repo *Repo) GetAll() []Article {
	articles := []Article{}

	rows, err := repo.DB.Query(`SELECT * FROM articles`)
	if err != nil {
		fmt.Println(err)
	}

	for rows.Next() {
		var article Article
		rows.Scan(&article.ID, &article.USER_ID, &article.Author,
			&article.Title, &article.Body, &article.Date)
		articles = append(articles, article)
	}

	return articles
}
