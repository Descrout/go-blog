package article

import (
	"database/sql"
	"log"
)

type Article struct {
	ID      int    `json: "id"`
	User_ID int    `json: "user_id"`
	Title   string `json: "title"`
	Body    string `json: "body"`
	Date    string `json: "date"`
}

type Repo struct {
	DB *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{
		DB: db,
	}
}

func (repo *Repo) Add(article Article) {
	stmt, err := repo.DB.Prepare(`
	INSERT INTO 
	articles (user_id,  title, body, date) 
	values (?, ?, ?, ?)`)
	if err != nil {
		log.Println(err)
	}
	defer stmt.Close()
	stmt.Exec(article.User_ID, article.Title, article.Body, article.Date)
}

func (repo *Repo) GetByID(id string) (*Article, error) {
	article := &Article{}
	stmt, err := repo.DB.Prepare("SELECT * FROM articles WHERE ID = ?")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(id).Scan(&article.ID, &article.User_ID,
		&article.Title, &article.Body, &article.Date)
	if err != nil {
		log.Println(err)
	}
	return article, err
}

func (repo *Repo) GetAll() []Article {
	articles := []Article{}
	rows, err := repo.DB.Query(`SELECT * FROM articles`)
	if err != nil {
		log.Println(err)
	}
	for rows.Next() {
		var article Article
		rows.Scan(&article.ID, &article.User_ID,
			&article.Title, &article.Body, &article.Date)
		articles = append(articles, article)
	}
	return articles
}
