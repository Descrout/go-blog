package article

import (
	"context"
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
		query: `SELECT id, user_id,
		title, created_at, updated_at,
		(SELECT COUNT(id) FROM favorites WHERE article_id = articles.id) fav_count,
		(SELECT COUNT(id) FROM comments WHERE article_id = articles.id) comment_count 
		FROM articles `,
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

func (s *Search) QueryFavoriteBy(userID string) {
	if userID != "" {
		s.ApplyCondition()
		s.query += `id IN (SELECT article_id FROM favorites WHERE user_id = ?) `
		s.params = append(s.params, userID)
	}
}

func (s *Search) QueryUserID(userID string) {
	if userID != "" {
		s.ApplyCondition()
		s.query += `user_id = ? `
		s.params = append(s.params, userID)
	}
}

func (s *Search) Limit(page int, sort string) {
	from := (page - 1) * ARTICLE_IN_PAGE

	switch sort {
	case "popular":
		s.query += `ORDER BY fav_count DESC, comment_count DESC `
	case "comment":
		s.query += `ORDER BY comment_count DESC, fav_count DESC `
	default:
		s.query += `ORDER BY created_at DESC `
	}

	s.query += `LIMIT ?, ?`
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

func (repo *Repo) ToggleFavoriteFor(id int64, userID int64) (bool, int, error) {
	ctx := context.Background()

	tx, err := repo.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return false, 0, err
	}

	row := tx.QueryRowContext(ctx, "SELECT 1 FROM favorites WHERE article_id = ? AND user_id = ?", id, userID)

	var favStatus bool

	switch err = row.Scan(&favStatus); err {
	case sql.ErrNoRows: // Did not favorited yet.
		_, err = tx.ExecContext(ctx, "INSERT INTO favorites (article_id, user_id) VALUES (?, ?)", id, userID)
		if err != nil {
			log.Println(err)
			tx.Rollback()
			return false, 0, err
		}
		favStatus = true
	case nil: // Already favorited.
		_, err = tx.ExecContext(ctx, "DELETE FROM favorites WHERE article_id = ? AND user_id = ?", id, userID)
		if err != nil {
			log.Println(err)
			tx.Rollback()
			return false, 0, err
		}
		favStatus = false
	default: // Query error
		log.Println(err)
		tx.Rollback()
		return false, 0, err
	}

	var favCount int
	err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM favorites WHERE article_id = ?", id).Scan(&favCount)
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return false, 0, err
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return false, 0, err
	}

	return favStatus, favCount, nil
}

func (repo *Repo) Delete(id int64) error {
	ctx := context.Background()

	tx, err := repo.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM articles WHERE id = ?", id)
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM favorites WHERE article_id = ?", id)
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM comments WHERE article_id = ?", id)
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
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

	stmt, err := repo.DB.Prepare(`SELECT *,
	(SELECT COUNT(id) FROM favorites WHERE article_id = articles.id) fav_count,
	(SELECT COUNT(id) FROM comments WHERE article_id = articles.id) comment_count 
	FROM articles WHERE id = ?`)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer stmt.Close()

	err = stmt.QueryRow(id).Scan(&article.ID, &article.User_ID,
		&article.Title, &article.Body, &article.Created_At, &article.Updated_At,
		&article.Favorites, &article.Comment_Count)

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
			&article.Title, &article.Created_At, &article.Updated_At,
			&article.Favorites, &article.Comment_Count)
		articles = append(articles, &article)
	}

	return articles
}
