package models

import (
	"os"
	"path/filepath"

	environment "github.com/nocodeleaks/quepasa/environment"
)

// CheckLocalSqliteExists checks if a local sqlite db file exists for the configured database name.
// It returns the filename (relative path) and true when exists, otherwise empty string and false.
func CheckLocalSqliteExists() (string, bool) {
	params := environment.Settings.Database.GetDBParameters()
	dbName := params.DataBase
	if dbName == "" {
		dbName = "quepasa"
	}

	c1 := dbName + ".db"
	c2 := dbName + ".sqlite"

	if _, err := os.Stat(c1); err == nil {
		p, _ := filepath.Abs(c1)
		return p, true
	}
	if _, err := os.Stat(c2); err == nil {
		p, _ := filepath.Abs(c2)
		return p, true
	}

	return "", false
}
