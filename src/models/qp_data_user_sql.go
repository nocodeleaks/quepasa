package models

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"

	log "github.com/sirupsen/logrus"
)

type QpDataUserSql struct {
	db *sqlx.DB
}

/*
	Count() (int, error)
	Exists(string) (bool, error)
	Find(string) (*QpUser, error)
	Check(string, password string) (*QpUser, error)
	Create(string, password string) (*QpUser, error)
*/

func (source QpDataUserSql) Count() (result int, err error) {
	err = source.db.Get(&result, "SELECT count(*) FROM users")
	return
}

func (source QpDataUserSql) Exists(username string) (bool, error) {
	sqlStmt := `SELECT username FROM users WHERE username = ?`
	err := source.db.QueryRow(sqlStmt, username).Scan(&username)
	if err != nil {
		if err != sql.ErrNoRows {
			return false, err
		}

		return false, nil
	}

	return true, nil
}

func (source QpDataUserSql) Find(username string) (result *QpUser, err error) {
	user := &QpUser{}
	err = source.db.Get(user, "SELECT username, password, timestamp FROM users WHERE username = ?", username)
	if err != nil {
		return
	}

	result = user
	return
}

func (source QpDataUserSql) Check(username string, password string) (result *QpUser, err error) {
	user := &QpUser{}
	err = source.db.Get(user, "SELECT username, password, timestamp FROM users WHERE username = ? LIMIT 1", username)
	if err != nil {
		return
	}

	if len(user.Username) == 0 {
		err = fmt.Errorf("user (%s) not found for check password", username)
		return
	} else {
		log.Infof("user hashed password: %s", user.Password)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return
	}

	result = user
	return
}

func (source QpDataUserSql) Create(username string, password string) (result *QpUser, err error) {
	if !ENV.AccountSetup() {
		err = fmt.Errorf("account creation is disabled by environment")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return
	}

	user := &QpUser{
		Username: username,
		Password: string(hashed),
	}

	query := `INSERT INTO users (username, password) VALUES (:username, :password)`
	_, err = source.db.NamedExec(query, user)
	if err != nil {
		return
	}

	result = user
	return
}

func (source QpDataUserSql) UpdatePassword(username string, password string) (err error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return
	}

	query := `UPDATE users SET password = ? WHERE username = ?`
	result, err := source.db.Exec(query, string(hashed), username)
	if err != nil {
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return
	}

	if affected == 0 {
		err = fmt.Errorf("user (%s) not found for update password", username)
	}

	return
}
