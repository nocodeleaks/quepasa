package environment

import (
	library "github.com/nocodeleaks/quepasa/library"
)

// Database environment variable names
const (
	ENV_DBDRIVER   = "DBDRIVER"   // database driver, default sqlite3
	ENV_DBHOST     = "DBHOST"     // database host
	ENV_DBDATABASE = "DBDATABASE" // database name
	ENV_DBPORT     = "DBPORT"     // database port
	ENV_DBUSER     = "DBUSER"     // database user
	ENV_DBPASSWORD = "DBPASSWORD" // database password
	ENV_DBSSLMODE  = "DBSSLMODE"  // database SSL mode
)

// DatabaseSettings holds all database configuration loaded from environment
type DatabaseSettings struct {
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	Database string `json:"database"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	SSLMode  string `json:"ssl_mode"`
}

// NewDatabaseSettings creates a new database settings by loading all values from environment
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

// GetDBParameters retrieves database connection parameters from the config.
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
