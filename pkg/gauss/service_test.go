package gauss

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/oauth2"
)

func TestGenerateStateUnique(t *testing.T) {
	svc, err := NewService("id", "secret", "http://example.com", "/dash", ScopeStrings(DefaultScopes), "")
	if err != nil {
		t.Fatalf("NewService error: %v", err)
	}
	a, err := svc.GenerateState()
	if err != nil {
		t.Fatalf("GenerateState error: %v", err)
	}
	b, err := svc.GenerateState()
	if err != nil {
		t.Fatalf("GenerateState error: %v", err)
	}
	if a == b {
		t.Errorf("expected unique states, got %s and %s", a, b)
	}
}

func TestGetUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"email":   "e@example.com",
			"name":    "tester",
			"picture": "img",
		})
	}))
	defer server.Close()

	orig := userInfoEndpoint
	userInfoEndpoint = server.URL
	defer func() { userInfoEndpoint = orig }()

	svc, err := NewService("id", "secret", "http://example.com", "/dash", ScopeStrings(DefaultScopes), "")
	if err != nil {
		t.Fatalf("NewService error: %v", err)
	}
	tok := &oauth2.Token{AccessToken: "abc"}
	user, err := svc.GetUser(tok)
	if err != nil {
		t.Fatalf("GetUser error: %v", err)
	}
	if user.Email != "e@example.com" || user.Name != "tester" {
		t.Fatalf("unexpected user: %+v", user)
	}
}
