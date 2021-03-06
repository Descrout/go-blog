package user

import (
	"context"
	"database/sql"
	"log"
)

const USERS_IN_PAGE = 10

type Search struct {
	query         string
	params        []interface{}
	isConditioned bool
}

func NewSearch() *Search {
	return &Search{
		query: `SELECT *,
		(SELECT COUNT(id) FROM favorites WHERE article_id IN (SELECT id FROM articles WHERE user_id = users.id)) karma 
		FROM users `,
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
		s.query += `name LIKE ? `
		s.params = append(s.params, keyword)
	}
}

func (s *Search) Limit(page int, popular bool) {
	from := (page - 1) * USERS_IN_PAGE

	if popular {
		s.query += `ORDER BY karma DESC, created_at DESC `
	} else {
		s.query += `ORDER BY created_at DESC `
	}

	s.query += `LIMIT ?, ?`
	s.params = append(s.params, from, USERS_IN_PAGE)
}

type Repo struct {
	DB *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{
		DB: db,
	}
}

func (repo *Repo) CheckCommentFor(id int64, articleID int64) bool {
	stmt, err := repo.DB.Prepare("SELECT 1 FROM comments WHERE article_id = ? AND user_id = ?")
	if err != nil {
		log.Println(err)
		return false
	}
	defer stmt.Close()

	var isCommented bool
	err = stmt.QueryRow(articleID, id).Scan(&isCommented)

	return err == nil
}

func (repo *Repo) CheckFavoriteFor(id int64, articleID int64) bool {
	stmt, err := repo.DB.Prepare("SELECT 1 FROM favorites WHERE article_id = ? AND user_id = ?")
	if err != nil {
		log.Println(err)
		return false
	}
	defer stmt.Close()

	var isFav bool
	err = stmt.QueryRow(articleID, id).Scan(&isFav)

	return err == nil
}

func (repo *Repo) Delete(id int64) error {
	ctx := context.Background()

	tx, err := repo.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM articles WHERE user_id = ?", id)
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM favorites WHERE user_id = ?", id)
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM comments WHERE user_id = ?", id)
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

func (repo *Repo) Update(id int64, field string, value interface{}) error {
	stmt, err := repo.DB.Prepare(`
	UPDATE users 
	SET ` + field + `= ?
	WHERE id = ?`)

	if err != nil {
		log.Println(err)
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(value, id)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (repo *Repo) DoesEmailExist(email string) (bool, error) {
	err := repo.DB.QueryRow("SELECT email FROM users WHERE email=?", email).Scan(&email)

	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		log.Println(err)
		return false, err
	}
	return true, nil
}

func (repo *Repo) Add(user *User) (int64, error) {
	stmt, err := repo.DB.Prepare(`
	INSERT INTO 
	users (name,  password, email, created_at) 
	values (?, ?, ?, ?)`)

	if err != nil {
		log.Println(err)
		return 0, err
	}

	defer stmt.Close()

	result, err := stmt.Exec(user.Name, user.Password, user.Email, user.Created_At)
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

func (repo *Repo) GetByEmail(email string) (*User, error) {
	user := &User{}

	stmt, err := repo.DB.Prepare("SELECT * FROM users WHERE email = ?")

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer stmt.Close()

	err = stmt.QueryRow(email).Scan(
		&user.ID, &user.Role_ID, &user.Name, &user.Password,
		&user.Email, &user.Image, &user.Created_At)

	if err != nil {
		log.Println(err)
	}

	return user, err
}

func (repo *Repo) GetByID(id int64) (*User, error) {
	user := &User{}

	stmt, err := repo.DB.Prepare("SELECT *,(SELECT COUNT(id) FROM favorites WHERE article_id IN (SELECT id FROM articles WHERE user_id = users.id)) karma  FROM users WHERE id = ?")

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer stmt.Close()

	err = stmt.QueryRow(id).Scan(
		&user.ID, &user.Role_ID, &user.Name, &user.Password,
		&user.Email, &user.Image, &user.Created_At, &user.Karma)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return user, err
}

func (repo *Repo) GetMultiple(search *Search) []*User {
	rows, err := repo.DB.Query(search.query, search.params...)
	if err != nil {
		log.Println(err)
	}

	users := []*User{}
	for rows.Next() {
		var user User
		rows.Scan(&user.ID, &user.Role_ID, &user.Name, &user.Password,
			&user.Email, &user.Image, &user.Created_At, &user.Karma)
		users = append(users, &user)
	}

	return users
}
