package comment

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

func (repo *Repo) GetByID(id string) (*Comment, error) {
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

func (repo *Repo) GetAllInArticle(articleID int64) []*Comment {
	comments := []*Comment{}

	rows, err := repo.DB.Query(`SELECT * FROM comments WHERE article_id=?`, articleID)

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
