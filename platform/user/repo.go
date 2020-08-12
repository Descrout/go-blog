package user

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

func (repo *Repo) Delete(id interface{}) error {
	stmt, err := repo.DB.Prepare("DELETE FROM users WHERE id = ?")

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

func (repo *Repo) Update(id interface{}, field string, value interface{}) error {
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
	users (name,  password, email) 
	values (?, ?, ?)`)

	if err != nil {
		log.Println(err)
		return 0, err
	}

	defer stmt.Close()

	result, err := stmt.Exec(user.Name, user.Password, user.Email)
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
		&user.Email, &user.Image)

	if err != nil {
		log.Println(err)
	}

	return user, err
}

func (repo *Repo) GetByID(id interface{}) (*User, error) {
	user := &User{}

	stmt, err := repo.DB.Prepare("SELECT * FROM users WHERE id = ?")

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer stmt.Close()

	err = stmt.QueryRow(id).Scan(
		&user.ID, &user.Role_ID, &user.Name, &user.Password,
		&user.Email, &user.Image)

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
			&user.Email, &user.Image)
		users = append(users, &user)
	}

	return users
}
