package session

import (
	"testing"
)

func TestStorePanicsWithoutInit(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic when store is nil")
		}
	}()
	Store()
}

func TestNewSessionAndStore(t *testing.T) {
	secret := []byte("secret")
	NewSession(secret)
	if Store() == nil {
		t.Fatal("store should not be nil after initialization")
	}
}
