package article

import (
	"database/sql"
	"log"
)

type Repo struct {
	DB *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{
		DB: db,
	}
}

func (repo *Repo) Add(article *Article) (int64, error) {
	stmt, err := repo.DB.Prepare(`
	INSERT INTO 
	articles (user_id,  title, body, date) 
	values (?, ?, ?, ?)`)

	if err != nil {
		log.Println(err)
		return 0, err
	}

	defer stmt.Close()

	result, err := stmt.Exec(article.User_ID, article.Title, article.Body, article.Date)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Println(err)
	}

	return id, err
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

func (repo *Repo) GetAll() []*Article {
	articles := []*Article{}

	rows, err := repo.DB.Query(`SELECT * FROM articles`)

	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		var article Article
		rows.Scan(&article.ID, &article.User_ID,
			&article.Title, &article.Body, &article.Date)
		articles = append(articles, &article)
	}

	return articles
}
