package environment

import (
	library "github.com/nocodeleaks/quepasa/library"
)

// Database environment variable names used to build the SQL connection for the
// Whatsmeow persistent store.
//
// Important: these variables are not the current source of truth for the
// internal QuePasa application database (`quepasa.sqlite` / `quepasa.db`).
const (
	ENV_DBDRIVER   = "DBDRIVER"   // SQL driver for the Whatsmeow store: sqlite3, postgres, or mysql. Default: sqlite3.
	ENV_DBHOST     = "DBHOST"     // Hostname for postgres/mysql. Ignored when DBDRIVER=sqlite3.
	ENV_DBDATABASE = "DBDATABASE" // Database name for postgres/mysql, or sqlite base file path/name for the Whatsmeow store.
	ENV_DBPORT     = "DBPORT"     // TCP port for postgres/mysql. Ignored when DBDRIVER=sqlite3.
	ENV_DBUSER     = "DBUSER"     // Username for postgres/mysql. Ignored when DBDRIVER=sqlite3.
	ENV_DBPASSWORD = "DBPASSWORD" // Password for postgres/mysql. Ignored when DBDRIVER=sqlite3.
	ENV_DBSSLMODE  = "DBSSLMODE"  // PostgreSQL sslmode value. Usually unused by sqlite3 and mysql.
)

// DatabaseSettings holds the SQL connection settings used by the Whatsmeow store
// started from `main.go`.
//
// For sqlite3, only Driver and Database are normally relevant.
// If Driver is sqlite3 and Database is empty, whatsmeow.Start() later falls back
// to the default database name `whatsmeow`.
type DatabaseSettings struct {
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	Database string `json:"database"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	SSLMode  string `json:"ssl_mode"`
}

// NewDatabaseSettings loads the Whatsmeow store connection settings from
// environment variables.
//
// These settings do not currently reconfigure the internal QuePasa application
// database used by models/migrations.
func NewDatabaseSettings() DatabaseSettings {
	return DatabaseSettings{
		Driver:   getEnvOrDefaultString(ENV_DBDRIVER, "sqlite3"),
		Host:     getEnvOrDefaultString(ENV_DBHOST, ""),
		Database: getEnvOrDefaultString(ENV_DBDATABASE, ""),
		Port:     getEnvOrDefaultString(ENV_DBPORT, ""),
		User:     getEnvOrDefaultString(ENV_DBUSER, ""),
		Password: getEnvOrDefaultString(ENV_DBPASSWORD, ""),
		SSLMode:  getEnvOrDefaultString(ENV_DBSSLMODE, ""),
	}
}

// GetDBParameters converts the environment-facing settings into the shared
// database parameter struct currently consumed by whatsmeow.Start().
func (config DatabaseSettings) GetDBParameters() library.DatabaseParameters {
	return library.DatabaseParameters{
		Driver:   config.Driver,
		Host:     config.Host,
		DataBase: config.Database,
		Port:     config.Port,
		User:     config.User,
		Password: config.Password,
		SSL:      config.SSLMode,
	}
}
