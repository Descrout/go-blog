package role

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
	stmt, err := repo.DB.Prepare("DELETE FROM roles WHERE id = ?")

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

func (repo *Repo) Update(role *Role) error {
	stmt, err := repo.DB.Prepare(`
	UPDATE roles 
	SET name = ?, code = ? 
	WHERE id = ?`)

	if err != nil {
		log.Println(err)
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(role.Name, role.Code, role.ID)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (repo *Repo) Add(role *Role) (int64, error) {
	stmt, err := repo.DB.Prepare(`
	INSERT INTO 
	roles (name,  code)
	values (?, ?)`)

	if err != nil {
		log.Println(err)
		return 0, err
	}

	defer stmt.Close()

	result, err := stmt.Exec(role.Name, role.Code)
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

func (repo *Repo) GetByID(id interface{}) (*Role, error) {
	role := &Role{}

	stmt, err := repo.DB.Prepare("SELECT * FROM roles WHERE id = ?")

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer stmt.Close()

	err = stmt.QueryRow(id).Scan(&role.ID, &role.Name, &role.Code)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return role, err
}

func (repo *Repo) GetAll() []*Role {
	roles := []*Role{}

	rows, err := repo.DB.Query(`SELECT * FROM roles`)

	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		var role Role
		rows.Scan(&role.ID, &role.Name, &role.Code)
		roles = append(roles, &role)
	}

	return roles
}
