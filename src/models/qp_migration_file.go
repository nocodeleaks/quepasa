package models

import (
	"io/ioutil"
	"os"

	"github.com/jmoiron/sqlx"
	migrate "github.com/joncalhoun/migrate"
)

type QpMigration struct {
	Id       string
	Title    string
	FileUp   string
	FileDown string
	Success  bool
}

func (source *QpMigration) MigrateTransaction() func(tx *sqlx.Tx) error {
	query, ok := FileToString(source.FileUp)
	if !ok {
		return nil
	}

	return func(tx *sqlx.Tx) error {
		_, err := tx.Exec(query)
		if err == nil {
			source.Success = true
		}
		return err
	}
}

func (source *QpMigration) RollbackTransaction() func(tx *sqlx.Tx) error {
	query, ok := FileToString(source.FileDown)
	if !ok {
		return nil
	}

	return func(tx *sqlx.Tx) error {
		_, err := tx.Exec(query)
		return err
	}
}

func FileToString(filename string) (string, bool) {
	if filename == "" {
		return "", false
	}
	f, err := os.Open(filename)
	if err != nil {
		// We could return a migration that errors when the migration is run, but I
		// think it makes more sense to panic here.
		panic(err)
	}
	fileBytes, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	return string(fileBytes), true
}

func (source *QpMigration) ToSqlxMigration() migrate.SqlxMigration {
	return migrate.SqlxMigration{
		ID:       source.Id,
		Migrate:  source.MigrateTransaction(),
		Rollback: source.RollbackTransaction(),
	}
}
