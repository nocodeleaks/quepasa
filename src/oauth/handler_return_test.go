package oauth

import (
	"net/http/httptest"
	"testing"
)

func TestNormalizeOAuthReturnURL(t *testing.T) {
	request := httptest.NewRequest("GET", "http://localhost:31000/oauth/login", nil)

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "relative app path", raw: "/apps/custom/", want: "/apps/custom/"},
		{name: "relative app path with query", raw: "/apps/custom/?tab=voice", want: "/apps/custom/?tab=voice"},
		{name: "same host absolute", raw: "http://localhost:31000/apps/custom/", want: "/apps/custom/"},
		{name: "loopback dev server", raw: "http://127.0.0.1:5174/apps/custom/", want: "http://127.0.0.1:5174/apps/custom/"},
		{name: "external host rejected", raw: "https://example.com/apps/custom/", want: ""},
		{name: "protocol relative rejected", raw: "//example.com/apps/custom/", want: ""},
		{name: "relative without slash rejected", raw: "apps/custom/", want: ""},
		{name: "javascript scheme rejected", raw: "javascript:alert(1)", want: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := normalizeOAuthReturnURL(request, test.raw)
			if got != test.want {
				t.Fatalf("normalizeOAuthReturnURL(%q) = %q, want %q", test.raw, got, test.want)
			}
		})
	}
}

func TestGetOAuthReturnURLPriority(t *testing.T) {
	request := httptest.NewRequest("GET", "http://localhost:31000/oauth/login?redirect=/apps/form/account&return_url=/apps/custom/&next=/", nil)

	got := getOAuthReturnURL(request)
	if got != "/apps/custom/" {
		t.Fatalf("getOAuthReturnURL() = %q, want /apps/custom/", got)
	}
}
