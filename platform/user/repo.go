package user

import (
	"context"
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

	stmt, err := repo.DB.Prepare("SELECT * FROM users WHERE id = ?")

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer stmt.Close()

	err = stmt.QueryRow(id).Scan(
		&user.ID, &user.Role_ID, &user.Name, &user.Password,
		&user.Email, &user.Image, &user.Created_At)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return user, err
}

func (repo *Repo) GetAll() []*User {
	users := []*User{}

	rows, err := repo.DB.Query(`SELECT * FROM users`)

	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		var user User
		rows.Scan(&user.ID, &user.Role_ID, &user.Name, &user.Password,
			&user.Email, &user.Image, &user.Created_At)
		users = append(users, &user)
	}

	return users
}
