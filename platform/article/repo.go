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

func (repo *Repo) Delete(id int64) error {
	stmt, err := repo.DB.Prepare("DELETE FROM articles WHERE id = ?")

	if err != nil {
		log.Println(err)
		return err
	}

	defer stmt.Close()

	if _, err = stmt.Exec(id); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (repo *Repo) Update(article *Article) error {
	stmt, err := repo.DB.Prepare("UPDATE articles SET title = ?, body = ?, updated_at = ? WHERE id = ?")

	if err != nil {
		log.Println(err)
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(article.Title, article.Body, article.Updated_At, article.ID)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (repo *Repo) Add(article *Article) (int64, error) {
	stmt, err := repo.DB.Prepare(`
	INSERT INTO 
	articles (user_id, title, body, created_at, updated_at) 
	values (?, ?, ?, ?, ?)`)

	if err != nil {
		log.Println(err)
		return 0, err
	}

	defer stmt.Close()

	result, err := stmt.Exec(article.User_ID, article.Title, article.Body, article.Created_At, article.Updated_At)
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
		&article.Title, &article.Body, &article.Created_At, &article.Updated_At)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return article, err
}

func (repo *Repo) GetAll() []*Article {
	articles := []*Article{}

	rows, err := repo.DB.Query(`SELECT id, user_id, title, created_at, updated_at FROM articles`)

	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		var article Article
		rows.Scan(&article.ID, &article.User_ID,
			&article.Title, &article.Created_At, &article.Updated_At)
		articles = append(articles, &article)
	}

	return articles
}
