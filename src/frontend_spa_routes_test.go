package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFrontendBetaCorePagesUseSPAEndpoints(t *testing.T) {
	t.Parallel()

	assertFileContains(t, filepath.Join("frontend", "src", "stores", "session.ts"), "/spa/session")
	assertFileNotContains(t, filepath.Join("frontend", "src", "stores", "session.ts"), "/api/session")

	assertFileContains(t, filepath.Join("frontend", "src", "App.vue"), "/spa/login/config")
	assertFileNotContains(t, filepath.Join("frontend", "src", "App.vue"), "/api/login/config")

	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Login.vue"), "/spa/login/config")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "Login.vue"), "/api/login/config")

	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Account.vue"), "/spa/session")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Account.vue"), "/spa/account")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Account.vue"), "/spa/account/masterkey")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "Account.vue"), "/api/account")

	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Environment.vue"), "/spa/environment")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "Environment.vue"), "/api/environment")

	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Users.vue"), "/spa/users")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Users.vue"), "/spa/user/")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "Users.vue"), "/api/users")

	assertFileContains(t, filepath.Join("frontend", "src", "pages", "UserCreate.vue"), "/spa/users")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "UserCreate.vue"), "/api/user")

	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Setup.vue"), "/spa/login/config")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Setup.vue"), "/spa/users")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "Setup.vue"), "/api/session")

	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Home.vue"), "/spa/servers")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Home.vue"), "/spa/servers/search")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Home.vue"), "/spa/server/create")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Home.vue"), "@/composables/useCableSubscription")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "Home.vue"), "/api/servers")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "Home.vue"), "signalr")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "Home.vue"), "@/services/cable")

	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Connect.vue"), "/spa/server/create")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "Connect.vue"), "/api/server/create")

	assertFileContains(t, filepath.Join("frontend", "src", "pages", "QRCode.vue"), "/spa/server/${token}/qrcode")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "QRCode.vue"), "/spa/server/${token}/info")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "QRCode.vue"), "@/composables/useCableSubscription")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "QRCode.vue"), "/api/server/${token}/qrcode")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "QRCode.vue"), "signalr")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "QRCode.vue"), "@/services/cable")

	assertFileContains(t, filepath.Join("frontend", "src", "pages", "PairCode.vue"), "/spa/server/${token}/paircode")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "PairCode.vue"), "/spa/server/${token}/info")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "PairCode.vue"), "@/composables/useCableSubscription")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "PairCode.vue"), "/api/server/${token}/paircode")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "PairCode.vue"), "signalr")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "PairCode.vue"), "@/services/cable")

	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Server.vue"), "/spa/server/${token}/info")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Server.vue"), "/spa/server/${token}/${endpoint}")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Server.vue"), "/spa/server/${token}")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Server.vue"), "@/composables/useCableSubscription")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "Server.vue"), "/api/server/")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "Server.vue"), "/api/command")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "Server.vue"), "signalr")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "Server.vue"), "@/services/cable")

	assertFileContains(t, filepath.Join("frontend", "src", "pages", "SendMessage.vue"), "/spa/server/${token}/contacts")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "SendMessage.vue"), "/spa/server/${token}/send")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "SendMessage.vue"), "/api/server/${token}/contacts")

	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Groups.vue"), "/spa/server/${token}/groups")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Groups.vue"), "/spa/server/${token}/groups/create")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Groups.vue"), "/spa/server/${token}/messages")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Groups.vue"), "/spa/server/${token}/picinfo/")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "Groups.vue"), "/api/groups/")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "Groups.vue"), "/api/server/")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "Groups.vue"), "/api/picinfo/")

	assertFileContains(t, filepath.Join("frontend", "src", "pages", "GroupDetail.vue"), "/spa/server/${token}/group/${encodedGroupId}")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "GroupDetail.vue"), "/spa/server/${token}/messages")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "GroupDetail.vue"), "/spa/server/${token}/info")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "GroupDetail.vue"), "/spa/server/${token}/picinfo/")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "GroupDetail.vue"), "/api/groups/")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "GroupDetail.vue"), "/api/server/")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "GroupDetail.vue"), "/api/picinfo/")

	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Webhooks.vue"), "/spa/server/${currentToken.value}/webhooks")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "Webhooks.vue"), "/api/webhooks")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "Webhooks.vue"), "/api/toggle")

	assertFileContains(t, filepath.Join("frontend", "src", "pages", "RabbitMQ.vue"), "/spa/server/${currentToken.value}/rabbitmq")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "RabbitMQ.vue"), "/api/rabbitmq")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "RabbitMQ.vue"), "/api/toggle")

	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Messages.vue"), "/spa/server/")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Messages.vue"), "/spa/server/${token}/messages")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Messages.vue"), "/spa/server/${token}/contacts")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Messages.vue"), "/spa/server/${token}/picinfo/")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Messages.vue"), "/spa/server/${token}/messages/${m.id}/history/download")
	assertFileContains(t, filepath.Join("frontend", "src", "pages", "Messages.vue"), "@/composables/useCableSubscription")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "Messages.vue"), "/api/server/")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "Messages.vue"), "/api/picinfo/")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "Messages.vue"), "../services/signalr")
	assertFileNotContains(t, filepath.Join("frontend", "src", "pages", "Messages.vue"), "../services/cable")

	assertFileContains(t, filepath.Join("frontend", "src", "services", "cable.ts"), "/cable")
	assertFileContains(t, filepath.Join("frontend", "src", "services", "cable.ts"), "'subscribe'")
	assertFileContains(t, filepath.Join("frontend", "src", "services", "cable.ts"), "retain()")
	assertFileContains(t, filepath.Join("frontend", "src", "services", "cable.ts"), "release()")
	assertFileContains(t, filepath.Join("frontend", "src", "composables", "useCableSubscription.ts"), "subscribeToken")
	assertFileContains(t, filepath.Join("frontend", "src", "composables", "useCableSubscription.ts"), "onConnectError")
	assertFileContains(t, filepath.Join("frontend", "src", "composables", "useCableSubscription.ts"), "cableService.retain()")
	assertFileContains(t, filepath.Join("frontend", "src", "composables", "useCableSubscription.ts"), "cableService.release()")
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
