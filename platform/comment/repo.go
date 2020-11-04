package comment

import (
	"database/sql"
	"log"
)

const COMMENTS_IN_PAGE = 10

type Search struct {
	query         string
	params        []interface{}
	isConditioned bool
}

func NewSearch() *Search {
	return &Search{
		query:         `SELECT * FROM comments `,
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
		s.query += `body LIKE ? `
		s.params = append(s.params, keyword)
	}
}

func (s *Search) QueryUserID(userID string) {
	if userID != "" {
		s.ApplyCondition()
		s.query += `user_id = ? `
		s.params = append(s.params, userID)
	}
}

func (s *Search) QueryArticleID(articleID int64) {
	s.ApplyCondition()
	s.query += `article_id = ? `
	s.params = append(s.params, articleID)
}

func (s *Search) Limit(page int) {
	from := (page - 1) * COMMENTS_IN_PAGE
	s.query += `ORDER BY created_at DESC LIMIT ?, ?`
	s.params = append(s.params, from, COMMENTS_IN_PAGE)
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
	stmt, err := repo.DB.Prepare("DELETE FROM comments WHERE id = ?")

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

func (repo *Repo) Update(comment *Comment) error {
	stmt, err := repo.DB.Prepare("UPDATE comments SET body = ?, updated_at = ? WHERE id = ?")

	if err != nil {
		log.Println(err)
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(comment.Body, comment.Updated_At, comment.ID)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (repo *Repo) Add(comment *Comment) (int64, error) {
	stmt, err := repo.DB.Prepare(`
	INSERT INTO 
	comments (user_id,  article_id, body, created_at, updated_at) 
	values (?, ?, ?, ?, ?)`)

	if err != nil {
		log.Println(err)
		return 0, err
	}

	defer stmt.Close()

	result, err := stmt.Exec(comment.User_ID, comment.Article_ID, comment.Body, comment.Created_At, comment.Updated_At)
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

func (repo *Repo) GetByID(id int64) (*Comment, error) {
	comment := &Comment{}

	stmt, err := repo.DB.Prepare("SELECT * FROM comments WHERE ID = ?")

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer stmt.Close()

	err = stmt.QueryRow(id).Scan(&comment.ID, &comment.User_ID,
		&comment.Article_ID, &comment.Body, &comment.Created_At, &comment.Updated_At)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return comment, err
}

func (repo *Repo) GetMultiple(search *Search) []*Comment {
	comments := []*Comment{}

	rows, err := repo.DB.Query(search.query, search.params...)

	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		var comment Comment
		rows.Scan(&comment.ID, &comment.User_ID,
			&comment.Article_ID, &comment.Body, &comment.Created_At, &comment.Updated_At)
		comments = append(comments, &comment)
	}

	return comments
}
