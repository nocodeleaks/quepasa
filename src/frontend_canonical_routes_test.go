package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFrontendBetaCorePagesUseCanonicalEndpoints(t *testing.T) {
	t.Parallel()

	assertFileContains(t, frontendSourcePath("stores", "session.ts"), "/api/auth/session")
	assertFileContains(t, frontendSourcePath("App.vue"), "/api/auth/config")
	assertFileContains(t, frontendSourcePath("pages", "Login.vue"), "/api/auth/config")
	assertFileContains(t, frontendSourcePath("pages", "Account.vue"), "/api/auth/session")
	assertFileContains(t, frontendSourcePath("pages", "Account.vue"), "/api/auth/account")
	assertFileContains(t, frontendSourcePath("pages", "Account.vue"), "/api/auth/masterkey/status")
	assertFileContains(t, frontendSourcePath("pages", "Environment.vue"), "/api/system/environment")
	assertFileContains(t, frontendSourcePath("pages", "Users.vue"), "/api/users")
	assertFileContains(t, frontendSourcePath("pages", "UserCreate.vue"), "/api/users")
	assertFileContains(t, frontendSourcePath("pages", "Setup.vue"), "/api/auth/config")
	assertFileContains(t, frontendSourcePath("pages", "Setup.vue"), "/api/users")
	assertFileContains(t, frontendSourcePath("pages", "Home.vue"), "/api/sessions")
	assertFileContains(t, frontendSourcePath("pages", "Home.vue"), "/api/sessions/search")
	assertFileContains(t, frontendSourcePath("pages", "Connect.vue"), "/api/sessions")
	assertFileContains(t, frontendSourcePath("pages", "QRCode.vue"), "/api/session/qrcode")
	assertFileContains(t, frontendSourcePath("pages", "QRCode.vue"), "/api/sessions/get")
	assertFileContains(t, frontendSourcePath("pages", "PairCode.vue"), "/api/session/paircode")
	assertFileContains(t, frontendSourcePath("pages", "PairCode.vue"), "/api/sessions/get")
	assertFileContains(t, frontendSourcePath("pages", "Server.vue"), "/api/sessions/get")
	assertFileContains(t, frontendSourcePath("pages", "Server.vue"), "/api/sessions")
	assertFileContains(t, frontendSourcePath("pages", "SendMessage.vue"), "/api/contacts")
	assertFileContains(t, frontendSourcePath("pages", "SendMessage.vue"), "/api/messages")
	assertFileContains(t, frontendSourcePath("pages", "Groups.vue"), "/api/groups")
	assertFileContains(t, frontendSourcePath("pages", "GroupDetail.vue"), "/api/groups/get")
	assertFileContains(t, frontendSourcePath("pages", "Webhooks.vue"), "/api/dispatches/webhooks")
	assertFileContains(t, frontendSourcePath("pages", "RabbitMQ.vue"), "/api/dispatches/rabbitmq")
	assertFileContains(t, frontendSourcePath("pages", "Messages.vue"), "/api/messages")
	assertFileContains(t, frontendSourcePath("pages", "Messages.vue"), "/api/contacts")
	assertFileContains(t, frontendSourcePath("pages", "Messages.vue"), "/api/media/download")
	assertFileContains(t, frontendSourcePath("pages", "Messages.vue"), "/api/chats/archive")
	assertFileContains(t, frontendSourcePath("pages", "Messages.vue"), "/api/chats/presence")

	assertFileContains(t, frontendSourcePath("services", "cable.ts"), "/cable")
	assertFileContains(t, frontendSourcePath("composables", "useCableSubscription.ts"), "subscribeToken")
	assertFileContains(t, frontendSourcePath("composables", "useServerLifecycleRefresh.ts"), "server.connected")
	assertFileContains(t, frontendSourcePath("composables", "useServerLifecycleRefresh.ts"), "server.deleted")
}

func frontendSourcePath(parts ...string) string {
	baseParts := []string{"apps", "vuejs", "client", "src"}
	return filepath.Join(append(baseParts, parts...)...)
}

func assertFileContains(t *testing.T, path string, want string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	if !strings.Contains(string(content), want) {
		t.Fatalf("%s does not contain %q", path, want)
	}
}
