package models

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type QpDataServerSql struct {
	db *sqlx.DB
}

func (source QpDataServerSql) FindForUser(token string, user string) (response *QpServer, err error) {
	err = source.db.Get(&response, "SELECT * FROM servers WHERE token = ? AND user = ?", token, user)
	return
}

func (source QpDataServerSql) FindAll() (response []*QpServer) {
	_ = source.db.Select(&response, "SELECT * FROM servers")
	return
}

func (source QpDataServerSql) Exists(token string) (bool, error) {
	sqlStmt := `SELECT token FROM servers WHERE token = ?`
	err := source.db.QueryRow(sqlStmt, token).Scan(&token)
	if err != nil {
		if err != sql.ErrNoRows {
			return false, err
		}

		return false, nil
	}

	return true, nil
}

func (source QpDataServerSql) FindByToken(token string) (response *QpServer, err error) {
	err = source.db.Get(&response, "SELECT * FROM servers WHERE token = ?", token)
	return
}

func (source QpDataServerSql) Add(element *QpServer) error {
	query := `INSERT INTO servers (token, wid, verified, devel, groups, broadcasts, readreceipts, calls, readupdate, user) VALUES (:token, :wid, :verified, :devel, :groups, :broadcasts, :readreceipts, :calls, :readupdate, :user)`
	params := map[string]any{
		"token":        element.Token,
		"wid":          nil,
		"verified":     element.Verified,
		"devel":        element.Devel,
		"groups":       element.Groups,
		"broadcasts":   element.Broadcasts,
		"readreceipts": element.ReadReceipts,
		"calls":        element.Calls,
		"readupdate":   element.ReadUpdate,
		"user":         element.User,
	}
	if len(element.Wid) > 0 {
		params["wid"] = element.Wid
	}
	_, err := source.db.NamedExec(query, params)
	return err
}

func (source QpDataServerSql) Update(element *QpServer) error {
	query := `UPDATE servers SET wid = :wid, verified = :verified, devel = :devel, groups = :groups, broadcasts = :broadcasts, readreceipts = :readreceipts, calls = :calls, readupdate = :readupdate, user = :user WHERE token = :token`
	params := map[string]any{
		"token":        element.Token,
		"wid":          nil,
		"verified":     element.Verified,
		"devel":        element.Devel,
		"groups":       element.Groups,
		"broadcasts":   element.Broadcasts,
		"readreceipts": element.ReadReceipts,
		"calls":        element.Calls,
		"readupdate":   element.ReadUpdate,
		"user":         element.User,
	}
	if len(element.Wid) > 0 {
		params["wid"] = element.Wid
	}
	_, err := source.db.NamedExec(query, params)
	return err
}

func (source QpDataServerSql) Delete(token string) error {
	query := `DELETE FROM servers WHERE token = ?`
	_, err := source.db.Exec(query, token)
	return err
}
