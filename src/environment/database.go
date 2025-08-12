package environment

import (
	library "github.com/nocodeleaks/quepasa/library"
)

// DatabaseEnvironment handles all database-related environment variables
type DatabaseEnvironment struct{}

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

// GetDBParameters retrieves database connection parameters from environment variables.
// It defaults to "sqlite3" if DBDRIVER is not set.
func (env *DatabaseEnvironment) GetDBParameters() library.DatabaseParameters {
	parameters := library.DatabaseParameters{}

	parameters.Driver = getEnvOrDefaultString(ENV_DBDRIVER, "sqlite3")
	parameters.Host = getEnvOrDefaultString(ENV_DBHOST, "")
	parameters.DataBase = getEnvOrDefaultString(ENV_DBDATABASE, "")
	parameters.Port = getEnvOrDefaultString(ENV_DBPORT, "")
	parameters.User = getEnvOrDefaultString(ENV_DBUSER, "")
	parameters.Password = getEnvOrDefaultString(ENV_DBPASSWORD, "")
	parameters.SSL = getEnvOrDefaultString(ENV_DBSSLMODE, "")
	return parameters
}

// Driver returns the database driver. Defaults to "sqlite3".
func (env *DatabaseEnvironment) Driver() string {
	return getEnvOrDefaultString(ENV_DBDRIVER, "sqlite3")
}

// Host returns the database host. Defaults to empty string.
func (env *DatabaseEnvironment) Host() string {
	return getEnvOrDefaultString(ENV_DBHOST, "")
}

// Database returns the database name. Defaults to empty string.
func (env *DatabaseEnvironment) Database() string {
	return getEnvOrDefaultString(ENV_DBDATABASE, "")
}

// Port returns the database port. Defaults to empty string.
func (env *DatabaseEnvironment) Port() string {
	return getEnvOrDefaultString(ENV_DBPORT, "")
}

// User returns the database user. Defaults to empty string.
func (env *DatabaseEnvironment) User() string {
	return getEnvOrDefaultString(ENV_DBUSER, "")
}

// Password returns the database password. Defaults to empty string.
func (env *DatabaseEnvironment) Password() string {
	return getEnvOrDefaultString(ENV_DBPASSWORD, "")
}

// SSLMode returns the database SSL mode. Defaults to empty string.
func (env *DatabaseEnvironment) SSLMode() string {
	return getEnvOrDefaultString(ENV_DBSSLMODE, "")
}
