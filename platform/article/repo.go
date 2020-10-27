package article

import (
	"database/sql"
	"log"
)

const ARTICLE_IN_PAGE = 10

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

func (repo *Repo) GetMultiple(page int, keyword string, dates [2]int64) []*Article {
	var rows *sql.Rows
	var err error
	pFrom, pTo := (page-1)*ARTICLE_IN_PAGE, ARTICLE_IN_PAGE

	if dates[1] > 0 {
		if keyword != "" {
			keyword = "%" + keyword + "%"
			rows, err = repo.DB.Query(
				`SELECT * FROM articles 
				WHERE created_at >= ? AND created_at <= ? 
				AND (title LIKE ? OR body LIKE ?) 
				ORDER BY created_at DESC 
				LIMIT ?, ?`,
				dates[0], dates[1],
				keyword, keyword,
				pFrom, pTo)
		} else {
			rows, err = repo.DB.Query(
				`SELECT * FROM articles 
				WHERE created_at >= ? AND created_at <= ? 
				ORDER BY created_at DESC 
				LIMIT ?, ?`,
				dates[0], dates[1],
				pFrom, pTo)
		}
	} else if keyword != "" {
		keyword = "%" + keyword + "%"

		rows, err = repo.DB.Query(
			`SELECT * FROM articles 
			WHERE title LIKE ? OR body LIKE ? 
			ORDER BY created_at DESC 
			LIMIT ?, ?`,
			keyword, keyword,
			pFrom, pTo)
	} else {
		rows, err = repo.DB.Query(`SELECT * FROM articles ORDER BY created_at DESC LIMIT ?, ?`,
			pFrom, pTo)
	}

	if err != nil {
		log.Println(err)
	}

	articles := []*Article{}
	for rows.Next() {
		var article Article
		rows.Scan(&article.ID, &article.User_ID,
			&article.Title, &article.Body, &article.Created_At, &article.Updated_At)
		articles = append(articles, &article)
	}

	return articles
}
