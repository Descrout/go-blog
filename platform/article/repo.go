package article

import (
	"database/sql"
	"log"
)

const ARTICLE_IN_PAGE = 10

type Search struct {
	query         string
	params        []interface{}
	isConditioned bool
}

func NewSearch() *Search {
	return &Search{
		query:         `SELECT * FROM articles `,
		params:        []interface{}{},
		isConditioned: false,
	}
}

func (s *Search) ApplyCondition() {
	if s.isConditioned {
		s.query += `AND `
	} else {
		s.query += `WHERE `
		s.isConditioned = true
	}
}

func (s *Search) QueryDate(from int64, to int64) {
	if to > 0 {
		s.ApplyCondition()
		s.query += `created_at >= ? AND created_at <= ? `
		s.params = append(s.params, from, to)
	}
}

func (s *Search) QueryKeyword(keyword string) {
	if keyword != "" {
		s.ApplyCondition()
		keyword = "%" + keyword + "%"
		s.query += `(title LIKE ? OR body LIKE ?) `
		s.params = append(s.params, keyword, keyword)
	}
}

func (s *Search) QueryUserID(userID string) {
	if userID != "" {
		s.ApplyCondition()
		s.query += `user_id = ? `
		s.params = append(s.params, userID)
	}
}

func (s *Search) Limit(page int) {
	from := (page - 1) * ARTICLE_IN_PAGE
	s.query += `ORDER BY created_at DESC LIMIT ?, ?`
	s.params = append(s.params, from, ARTICLE_IN_PAGE)
}

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

func (repo *Repo) GetMultiple(search *Search) []*Article {
	rows, err := repo.DB.Query(search.query, search.params...)
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
