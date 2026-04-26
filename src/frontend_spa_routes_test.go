package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFrontendBetaCorePagesUseSPAEndpoints(t *testing.T) {
	t.Parallel()

	assertFileContains(t, spaSourcePath("stores", "session.ts"), "/spa/session")
	assertFileNotContains(t, spaSourcePath("stores", "session.ts"), "/api/session")

	assertFileContains(t, spaSourcePath("App.vue"), "/spa/login/config")
	assertFileNotContains(t, spaSourcePath("App.vue"), "/api/login/config")

	assertFileContains(t, spaSourcePath("pages", "Login.vue"), "/spa/login/config")
	assertFileNotContains(t, spaSourcePath("pages", "Login.vue"), "/api/login/config")

	assertFileContains(t, spaSourcePath("pages", "Account.vue"), "/spa/session")
	assertFileContains(t, spaSourcePath("pages", "Account.vue"), "/spa/account")
	assertFileContains(t, spaSourcePath("pages", "Account.vue"), "/spa/account/masterkey")
	assertFileNotContains(t, spaSourcePath("pages", "Account.vue"), "/api/account")

	assertFileContains(t, spaSourcePath("pages", "Environment.vue"), "/spa/environment")
	assertFileNotContains(t, spaSourcePath("pages", "Environment.vue"), "/api/environment")

	assertFileContains(t, spaSourcePath("pages", "Users.vue"), "/spa/users")
	assertFileContains(t, spaSourcePath("pages", "Users.vue"), "/spa/user/")
	assertFileNotContains(t, spaSourcePath("pages", "Users.vue"), "/api/users")

	assertFileContains(t, spaSourcePath("pages", "UserCreate.vue"), "/spa/users")
	assertFileNotContains(t, spaSourcePath("pages", "UserCreate.vue"), "/api/user")

	assertFileContains(t, spaSourcePath("pages", "Setup.vue"), "/spa/login/config")
	assertFileContains(t, spaSourcePath("pages", "Setup.vue"), "/spa/users")
	assertFileNotContains(t, spaSourcePath("pages", "Setup.vue"), "/api/session")

	assertFileContains(t, spaSourcePath("pages", "Home.vue"), "/spa/servers")
	assertFileContains(t, spaSourcePath("pages", "Home.vue"), "/spa/servers/search")
	assertFileContains(t, spaSourcePath("pages", "Home.vue"), "/spa/server/create")
	assertFileContains(t, spaSourcePath("pages", "Home.vue"), "@/composables/useServerLifecycleRefresh")
	assertFileNotContains(t, spaSourcePath("pages", "Home.vue"), "/api/servers")
	assertFileNotContains(t, spaSourcePath("pages", "Home.vue"), "signalr")
	assertFileNotContains(t, spaSourcePath("pages", "Home.vue"), "@/services/cable")
	assertFileNotContains(t, spaSourcePath("pages", "Home.vue"), "@/composables/useCableSubscription")

	assertFileContains(t, spaSourcePath("pages", "Connect.vue"), "/spa/server/create")
	assertFileNotContains(t, spaSourcePath("pages", "Connect.vue"), "/api/server/create")

	assertFileContains(t, spaSourcePath("pages", "QRCode.vue"), "/spa/server/${token}/qrcode")
	assertFileContains(t, spaSourcePath("pages", "QRCode.vue"), "/spa/server/${token}/info")
	assertFileContains(t, spaSourcePath("pages", "QRCode.vue"), "@/composables/useCableSubscription")
	assertFileNotContains(t, spaSourcePath("pages", "QRCode.vue"), "/api/server/${token}/qrcode")
	assertFileNotContains(t, spaSourcePath("pages", "QRCode.vue"), "signalr")
	assertFileNotContains(t, spaSourcePath("pages", "QRCode.vue"), "@/services/cable")
	assertFileNotContains(t, spaSourcePath("pages", "QRCode.vue"), "setInterval(")

	assertFileContains(t, spaSourcePath("pages", "PairCode.vue"), "/spa/server/${token}/paircode")
	assertFileContains(t, spaSourcePath("pages", "PairCode.vue"), "/spa/server/${token}/info")
	assertFileContains(t, spaSourcePath("pages", "PairCode.vue"), "@/composables/useCableSubscription")
	assertFileNotContains(t, spaSourcePath("pages", "PairCode.vue"), "/api/server/${token}/paircode")
	assertFileNotContains(t, spaSourcePath("pages", "PairCode.vue"), "signalr")
	assertFileNotContains(t, spaSourcePath("pages", "PairCode.vue"), "@/services/cable")
	assertFileNotContains(t, spaSourcePath("pages", "PairCode.vue"), "setInterval(")

	assertFileContains(t, spaSourcePath("pages", "Server.vue"), "/spa/server/${token}/info")
	assertFileContains(t, spaSourcePath("pages", "Server.vue"), "/spa/server/${token}/${endpoint}")
	assertFileContains(t, spaSourcePath("pages", "Server.vue"), "/spa/server/${token}")
	assertFileContains(t, spaSourcePath("pages", "Server.vue"), "@/composables/useServerLifecycleRefresh")
	assertFileNotContains(t, spaSourcePath("pages", "Server.vue"), "/api/server/")
	assertFileNotContains(t, spaSourcePath("pages", "Server.vue"), "/api/command")
	assertFileNotContains(t, spaSourcePath("pages", "Server.vue"), "signalr")
	assertFileNotContains(t, spaSourcePath("pages", "Server.vue"), "@/services/cable")
	assertFileNotContains(t, spaSourcePath("pages", "Server.vue"), "@/composables/useCableSubscription")

	assertFileContains(t, spaSourcePath("pages", "SendMessage.vue"), "/spa/server/${token}/contacts")
	assertFileContains(t, spaSourcePath("pages", "SendMessage.vue"), "/spa/server/${token}/send")
	assertFileNotContains(t, spaSourcePath("pages", "SendMessage.vue"), "/api/server/${token}/contacts")

	assertFileContains(t, spaSourcePath("pages", "Groups.vue"), "/spa/server/${token}/groups")
	assertFileContains(t, spaSourcePath("pages", "Groups.vue"), "/spa/server/${token}/groups/create")
	assertFileContains(t, spaSourcePath("pages", "Groups.vue"), "/spa/server/${token}/messages")
	assertFileContains(t, spaSourcePath("pages", "Groups.vue"), "/spa/server/${token}/picinfo/")
	assertFileNotContains(t, spaSourcePath("pages", "Groups.vue"), "/api/groups/")
	assertFileNotContains(t, spaSourcePath("pages", "Groups.vue"), "/api/server/")
	assertFileNotContains(t, spaSourcePath("pages", "Groups.vue"), "/api/picinfo/")

	assertFileContains(t, spaSourcePath("pages", "GroupDetail.vue"), "/spa/server/${token}/group/${encodedGroupId}")
	assertFileContains(t, spaSourcePath("pages", "GroupDetail.vue"), "/spa/server/${token}/messages")
	assertFileContains(t, spaSourcePath("pages", "GroupDetail.vue"), "/spa/server/${token}/info")
	assertFileContains(t, spaSourcePath("pages", "GroupDetail.vue"), "/spa/server/${token}/picinfo/")
	assertFileNotContains(t, spaSourcePath("pages", "GroupDetail.vue"), "/api/groups/")
	assertFileNotContains(t, spaSourcePath("pages", "GroupDetail.vue"), "/api/server/")
	assertFileNotContains(t, spaSourcePath("pages", "GroupDetail.vue"), "/api/picinfo/")

	assertFileContains(t, spaSourcePath("pages", "Webhooks.vue"), "/spa/server/${currentToken.value}/webhooks")
	assertFileNotContains(t, spaSourcePath("pages", "Webhooks.vue"), "/api/webhooks")
	assertFileNotContains(t, spaSourcePath("pages", "Webhooks.vue"), "/api/toggle")

	assertFileContains(t, spaSourcePath("pages", "RabbitMQ.vue"), "/spa/server/${currentToken.value}/rabbitmq")
	assertFileNotContains(t, spaSourcePath("pages", "RabbitMQ.vue"), "/api/rabbitmq")
	assertFileNotContains(t, spaSourcePath("pages", "RabbitMQ.vue"), "/api/toggle")

	assertFileContains(t, spaSourcePath("pages", "Messages.vue"), "/spa/server/")
	assertFileContains(t, spaSourcePath("pages", "Messages.vue"), "/spa/server/${token}/messages")
	assertFileContains(t, spaSourcePath("pages", "Messages.vue"), "/spa/server/${token}/contacts")
	assertFileContains(t, spaSourcePath("pages", "Messages.vue"), "/spa/server/${token}/picinfo/")
	assertFileContains(t, spaSourcePath("pages", "Messages.vue"), "/spa/server/${token}/messages/${m.id}/history/download")
	assertFileContains(t, spaSourcePath("pages", "Messages.vue"), "@/composables/useCableSubscription")
	assertFileNotContains(t, spaSourcePath("pages", "Messages.vue"), "/api/server/")
	assertFileNotContains(t, spaSourcePath("pages", "Messages.vue"), "/api/picinfo/")
	assertFileNotContains(t, spaSourcePath("pages", "Messages.vue"), "../services/signalr")
	assertFileNotContains(t, spaSourcePath("pages", "Messages.vue"), "../services/cable")

	assertFileContains(t, spaSourcePath("services", "cable.ts"), "/cable")
	assertFileContains(t, spaSourcePath("services", "cable.ts"), "'subscribe'")
	assertFileContains(t, spaSourcePath("services", "cable.ts"), "retain()")
	assertFileContains(t, spaSourcePath("services", "cable.ts"), "release()")
	assertFileContains(t, spaSourcePath("composables", "useCableSubscription.ts"), "subscribeToken")
	assertFileContains(t, spaSourcePath("composables", "useCableSubscription.ts"), "onConnectError")
	assertFileContains(t, spaSourcePath("composables", "useCableSubscription.ts"), "cableService.retain()")
	assertFileContains(t, spaSourcePath("composables", "useCableSubscription.ts"), "cableService.release()")
	assertFileContains(t, spaSourcePath("composables", "useServerLifecycleRefresh.ts"), "server.connected")
	assertFileContains(t, spaSourcePath("composables", "useServerLifecycleRefresh.ts"), "server.deleted")
	assertFileContains(t, spaSourcePath("composables", "useServerLifecycleRefresh.ts"), "onDeleted")
}

func spaSourcePath(parts ...string) string {
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

func assertFileNotContains(t *testing.T, path string, unwanted string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	if strings.Contains(string(content), unwanted) {
		t.Fatalf("%s still contains %q", path, unwanted)
	}
}
