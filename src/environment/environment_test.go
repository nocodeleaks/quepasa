package environment

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

// init loads .env file before running tests
func init() {
	// Try to load .env file if it exists (won't fail if file doesn't exist)
	_ = godotenv.Load()
}

// TestDotEnvLoading tests if .env file is being loaded correctly
func TestDotEnvLoading(t *testing.T) {
	t.Log("🔍 Testing .env file loading")

	// Test key variables that should be in .env
	testVars := map[string]string{
		"WEBAPIPORT":    "31000",
		"APP_TITLE":     "Quepasa",
		"SIPPROXY_HOST": "voip.sufficit.com.br",
		"MASTERKEY":     "uiuiui",
		"GROUPS":        "true",
		"CALLS":         "false",
	}

	loadedCount := 0
	for key, expectedValue := range testVars {
		actualValue := os.Getenv(key)
		if actualValue != "" {
			loadedCount++
			if actualValue == expectedValue {
				t.Logf("✅ %s: '%s' (correct)", key, actualValue)
			} else {
				t.Logf("⚠️  %s: '%s' (expected: '%s')", key, actualValue, expectedValue)
			}
		} else {
			t.Logf("❌ %s: not loaded", key)
		}
	}

	if loadedCount == 0 {
		t.Error("❌ No environment variables loaded from .env file")
		t.Log("💡 Check if .env file exists and is readable")
	} else {
		t.Logf("✅ Loaded %d/%d environment variables from .env", loadedCount, len(testVars))
	}

	// Test SIP Proxy specifically
	sipHost := os.Getenv("SIPPROXY_HOST")
	if sipHost == "voip.sufficit.com.br" {
		t.Log("🎯 SIP Proxy should be ENABLED with loaded .env")
	} else {
		t.Logf("⚠️  SIP Proxy host: '%s' (expected: 'voip.sufficit.com.br')", sipHost)
	}
}

// TestEnvironmentPackageStructure tests if all environment files exist
func TestEnvironmentPackageStructure(t *testing.T) {
	expectedFiles := []string{
		"environment.go",
		"api.go",
		"database.go",
		"whatsapp.go",
		"sipproxy.go",
		"logging.go",
		"general.go",
		"rabbitmq.go",
		"README.md",
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected environment file not found: %s", file)
		}
	}
}

// TestEnvironmentVariablesDefault tests default values when no environment variables are set
func TestEnvironmentVariablesDefault(t *testing.T) {
	// Clear environment variables for default testing
	originalEnvs := make(map[string]string)
	testVars := []string{"WEBAPIPORT", "DBDRIVER", "SIPPROXY_HOST", "LOGLEVEL"}

	// Store original values
	for _, key := range testVars {
		originalEnvs[key] = os.Getenv(key)
		os.Unsetenv(key)
	}

	// Test default values
	if Settings.API.Port() != "31000" {
		t.Errorf("Expected default API port 31000, got %s", Settings.API.Port())
	}

	if Settings.Database.GetDBParameters().Driver != "sqlite3" {
		t.Errorf("Expected default database driver sqlite3, got %s", Settings.Database.GetDBParameters().Driver)
	}

	if Settings.SIPProxy.Enabled() {
		t.Error("SIP Proxy should be disabled when SIPPROXY_HOST is not set")
	}

	// Restore original values
	for key, value := range originalEnvs {
		if value != "" {
			os.Setenv(key, value)
		}
	}
}

// TestEnvironmentVariablesFromSystem tests loading from actual system environment
func TestEnvironmentVariablesFromSystem(t *testing.T) {
	t.Log("Testing environment variables from current system/file configuration")

	// Test API Configuration
	apiPort := Settings.API.Port()
	if apiPort == "" {
		t.Error("API Port should not be empty")
	}
	t.Logf("✅ API Port: %s", apiPort)

	// Test Database Configuration
	dbParams := Settings.Database.GetDBParameters()
	if dbParams.Driver == "" {
		t.Error("Database driver should not be empty")
	}
	t.Logf("✅ Database Driver: %s", dbParams.Driver)

	// Test SIP Proxy Configuration
	sipEnabled := Settings.SIPProxy.Enabled()
	sipHost := Settings.SIPProxy.Host()

	if sipEnabled {
		if sipHost == "" {
			t.Error("SIP Proxy is enabled but host is empty")
		}
		t.Logf("✅ SIP Proxy ENABLED - Host: %s", sipHost)
		t.Logf("   Port: %d", Settings.SIPProxy.Port())
		t.Logf("   STUN: %s", Settings.SIPProxy.STUNServer())
	} else {
		t.Log("✅ SIP Proxy DISABLED (SIPPROXY_HOST not set)")
	}

	// Test WhatsApp Configuration
	groups := Settings.WhatsApp.Groups()
	t.Logf("✅ WhatsApp Groups: %v", groups)

	calls := Settings.WhatsApp.Calls()
	t.Logf("✅ WhatsApp Calls: %v", calls)

	// Test General Configuration
	appTitle := Settings.General.AppTitle()
	t.Logf("✅ App Title: '%s'", appTitle)

	synopsisLength := Settings.General.SynopsisLength()
	if synopsisLength == 0 {
		t.Error("Synopsis length should not be zero")
	}
	t.Logf("✅ Synopsis Length: %d", synopsisLength)

	// Test Logging Configuration
	logLevel := Settings.Logging.LogLevel()
	t.Logf("✅ Log Level: '%s'", logLevel)

	// Test RabbitMQ Configuration
	rabbitmqQueue := Settings.RabbitMQ.Queue()
	rabbitmqConn := Settings.RabbitMQ.ConnectionString()

	if rabbitmqQueue != "" && rabbitmqConn != "" {
		t.Logf("✅ RabbitMQ ENABLED - Queue: %s", rabbitmqQueue)
	} else {
		t.Log("✅ RabbitMQ DISABLED")
	}
}

// TestSIPProxyActivationLogic tests SIP proxy activation based on HOST configuration
func TestSIPProxyActivationLogic(t *testing.T) {
	t.Log("Testing SIP Proxy activation logic")

	// Store original value
	originalHost := os.Getenv("SIPPROXY_HOST")

	// Test 1: SIP Proxy should be disabled when HOST is empty
	os.Unsetenv("SIPPROXY_HOST")
	if Settings.SIPProxy.Enabled() {
		t.Error("SIP Proxy should be disabled when SIPPROXY_HOST is empty")
	}
	t.Log("✅ SIP Proxy correctly disabled when HOST is empty")

	// Test 2: SIP Proxy should be enabled when HOST is set
	os.Setenv("SIPPROXY_HOST", "test.example.com")
	if !Settings.SIPProxy.Enabled() {
		t.Error("SIP Proxy should be enabled when SIPPROXY_HOST is set")
	}
	if Settings.SIPProxy.Host() != "test.example.com" {
		t.Errorf("Expected SIP host 'test.example.com', got '%s'", Settings.SIPProxy.Host())
	}
	t.Log("✅ SIP Proxy correctly enabled when HOST is set")

	// Restore original value
	if originalHost != "" {
		os.Setenv("SIPPROXY_HOST", originalHost)
	} else {
		os.Unsetenv("SIPPROXY_HOST")
	}
}

// TestEnvironmentSettingsSingleton tests if Settings singleton is properly initialized
func TestEnvironmentSettingsSingleton(t *testing.T) {
	if Settings.API == nil {
		t.Error("Settings.API should not be nil")
	}

	if Settings.Database == nil {
		t.Error("Settings.Database should not be nil")
	}

	if Settings.SIPProxy == nil {
		t.Error("Settings.SIPProxy should not be nil")
	}

	if Settings.WhatsApp == nil {
		t.Error("Settings.WhatsApp should not be nil")
	}

	if Settings.General == nil {
		t.Error("Settings.General should not be nil")
	}

	if Settings.Logging == nil {
		t.Error("Settings.Logging should not be nil")
	}

	if Settings.RabbitMQ == nil {
		t.Error("Settings.RabbitMQ should not be nil")
	}

	t.Log("✅ All Settings components properly initialized")
}

// TestEnvironmentVariablesCoverage tests if all expected environment variables are covered
func TestEnvironmentVariablesCoverage(t *testing.T) {
	// Test if key methods exist and don't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Environment method panicked: %v", r)
		}
	}()

	// API Environment
	_ = Settings.API.Port()
	_ = Settings.API.Host()
	_ = Settings.API.UseSSLForWebSocket()
	_ = Settings.API.SigningSecret()
	_ = Settings.API.MasterKey()
	_ = Settings.API.HTTPLogs()

	// Database Environment
	_ = Settings.Database.GetDBParameters()

	// SIP Proxy Environment
	_ = Settings.SIPProxy.Enabled()
	_ = Settings.SIPProxy.Host()
	_ = Settings.SIPProxy.Port()
	_ = Settings.SIPProxy.LocalPort()
	_ = Settings.SIPProxy.PublicIP()
	_ = Settings.SIPProxy.STUNServer()
	_ = Settings.SIPProxy.UseUPnP()
	_ = Settings.SIPProxy.MediaPorts()
	_ = Settings.SIPProxy.Codecs()
	_ = Settings.SIPProxy.UserAgent()
	_ = Settings.SIPProxy.LogLevel()
	_ = Settings.SIPProxy.Timeout()
	_ = Settings.SIPProxy.Retries()

	// WhatsApp Environment
	_ = Settings.WhatsApp.ReadUpdate()
	_ = Settings.WhatsApp.ReadReceipts()
	_ = Settings.WhatsApp.Calls()
	_ = Settings.WhatsApp.Groups()
	_ = Settings.WhatsApp.Broadcasts()
	_ = Settings.WhatsApp.HistorySync()
	_ = Settings.WhatsApp.Presence()

	// General Environment
	_ = Settings.General.Migrate()
	_ = Settings.General.MigrationPath()
	_ = Settings.General.AppTitle()
	_ = Settings.General.ShouldRemoveDigit9()
	_ = Settings.General.SynopsisLength()
	_ = Settings.General.CacheLength()
	_ = Settings.General.CacheDays()
	_ = Settings.General.UseCompatibleMIMEsAsAudio()
	_ = Settings.General.Testing()
	_ = Settings.General.AccountSetup()
	_ = Settings.General.DispatchUnhandled()

	// Logging Environment
	_ = Settings.Logging.LogLevel()
	_ = Settings.Logging.WhatsmeowLogLevel()
	_ = Settings.Logging.WhatsmeowDBLogLevel()

	// RabbitMQ Environment
	_ = Settings.RabbitMQ.Queue()
	_ = Settings.RabbitMQ.ConnectionString()
	_ = Settings.RabbitMQ.CacheLength()

	t.Log("✅ All environment methods accessible without panic")
}
