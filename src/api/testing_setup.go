package api

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	environment "github.com/nocodeleaks/quepasa/environment"
	models "github.com/nocodeleaks/quepasa/models"
)

var (
	testDB         *sqlx.DB
	testDBInitDone = false
)

// SetupTestDatabase creates an in-memory SQLite database for testing
func SetupTestDatabase(t *testing.T) *sqlx.DB {
	t.Helper()

	// Use in-memory database for tests
	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create schema
	if err := createTestSchema(db); err != nil {
		db.Close()
		t.Fatalf("Failed to create test schema: %v", err)
	}

	testDB = db
	testDBInitDone = true

	return db
}

// createTestSchema creates the necessary tables for testing
func createTestSchema(db *sqlx.DB) error {
	schema := `
		-- Users table
		CREATE TABLE IF NOT EXISTS users (
			username TEXT PRIMARY KEY,
			password TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		-- Servers table
		CREATE TABLE IF NOT EXISTS servers (
			token TEXT PRIMARY KEY,
			wid TEXT,
			user TEXT NOT NULL,
			verified BOOLEAN DEFAULT 0,
			devel BOOLEAN DEFAULT 0,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			groups INTEGER DEFAULT 1,
			direct INTEGER DEFAULT 1,
			broadcasts INTEGER DEFAULT 1,
			readreceipts INTEGER DEFAULT 1,
			calls INTEGER DEFAULT 1,
			readupdate INTEGER DEFAULT 1,
			FOREIGN KEY (user) REFERENCES users(username)
		);

		-- Dispatching table (webhooks and rabbitmq)
		CREATE TABLE IF NOT EXISTS dispatching (
			context TEXT NOT NULL,
			connection_string TEXT NOT NULL,
			type TEXT NOT NULL DEFAULT 'webhook',
			forwardinternal BOOLEAN DEFAULT 0,
			trackid TEXT,
			extra TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			groups INTEGER DEFAULT 1,
			direct INTEGER DEFAULT 1,
			broadcasts INTEGER DEFAULT 1,
			readreceipts INTEGER DEFAULT 1,
			calls INTEGER DEFAULT 1,
			PRIMARY KEY (context, connection_string)
		);
	`

	_, err := db.Exec(schema)
	return err
}

// SetupTestService initializes WhatsappService with test database
func SetupTestService(t *testing.T) {
	t.Helper()

	if !testDBInitDone {
		SetupTestDatabase(t)
	}

	// Initialize WhatsappService if not already done
	if models.WhatsappService == nil {
		models.WhatsappService = &models.QPWhatsappService{
			Servers: make(map[string]*models.QpWhatsappServer),
			DB: &models.QpDatabase{
				Connection: testDB,
			},
		}
	}

	// Initialize database interfaces using exported constructor functions
	if models.WhatsappService.DB != nil {
		// Initialize Users interface
		models.WhatsappService.DB.Users = models.NewQpDataUserSql(testDB)

		// Initialize Dispatching interface
		models.WhatsappService.DB.Dispatching = models.NewQpDataServerDispatchingSql(testDB)

		// Initialize Servers interface
		models.WhatsappService.DB.Servers = models.NewQpDataServerSql(testDB)
	}
}

// CleanupTestDatabase closes and cleans up the test database
func CleanupTestDatabase(t *testing.T) {
	t.Helper()

	if testDB != nil {
		testDB.Close()
		testDB = nil
		testDBInitDone = false
	}

	// Clear WhatsappService servers
	if models.WhatsappService != nil && models.WhatsappService.Servers != nil {
		models.WhatsappService.Servers = make(map[string]*models.QpWhatsappServer)
	}
}

// CreateTestUser creates a test user in the database
func CreateTestUser(t *testing.T, username, password string) *models.QpUser {
	t.Helper()

	user, err := models.WhatsappService.DB.Users.Create(username, password)
	if err != nil {
		t.Fatalf("Failed to create test user %s: %v", username, err)
	}

	return user
}

// CreateTestServer creates a test server in memory (not in DB)
func CreateTestServer(t *testing.T, token, username string) *models.QpWhatsappServer {
	t.Helper()

	serverInfo := &models.QpServer{
		Token: token,
		User:  username,
	}

	server, err := models.WhatsappService.AppendNewServer(serverInfo)
	if err != nil {
		t.Fatalf("Failed to create test server %s: %v", token, err)
	}

	return server
}

// SetupTestMasterKey sets a test master key for tests
func SetupTestMasterKey(t *testing.T, masterKey string) func() {
	t.Helper()

	oldMasterKey := environment.Settings.API.MasterKey
	environment.Settings.API.MasterKey = masterKey

	// Return cleanup function
	return func() {
		environment.Settings.API.MasterKey = oldMasterKey
	}
}

// GetTestDataDir returns the directory for test data files
func GetTestDataDir(t *testing.T) string {
	t.Helper()

	// Create temp directory for test files
	tmpDir := filepath.Join(os.TempDir(), "quepasa_tests")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}

	return tmpDir
}

// CreateTestDatabase creates a file-based SQLite database for tests that need persistence
func CreateTestDatabase(t *testing.T, name string) *sqlx.DB {
	t.Helper()

	testDir := GetTestDataDir(t)
	dbPath := filepath.Join(testDir, fmt.Sprintf("%s.db", name))

	// Remove existing test database
	os.Remove(dbPath)

	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database file: %v", err)
	}

	if err := createTestSchema(db); err != nil {
		db.Close()
		t.Fatalf("Failed to create test schema: %v", err)
	}

	// Register cleanup
	t.Cleanup(func() {
		db.Close()
		os.Remove(dbPath)
	})

	return db
}
