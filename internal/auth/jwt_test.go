package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// go test -v ./internal/auth/...

func TestMakeAndValidateJWT(t *testing.T) {
	secret := "super-secret-key"
	userID := uuid.New()

	// 1. Test successful creation and validation
	tokenString, err := MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error making JWT: %v", err)
	}

	parsedID, err := ValidateJWT(tokenString, secret)
	if err != nil {
		t.Fatalf("unexpected error validating JWT: %v", err)
	}

	if parsedID != userID {
		t.Errorf("expected user ID %v, got %v", userID, parsedID)
	}

	// 2. Test expired token rejection
	expiredToken, err := MakeJWT(userID, secret, -time.Hour) // expired 1 hour ago
	if err != nil {
		t.Fatalf("unexpected error making expired JWT: %v", err)
	}

	_, err = ValidateJWT(expiredToken, secret)
	if err == nil {
		t.Error("expected error for expired token, got nil")
	}

	// 3. Test wrong secret rejection
	_, err = ValidateJWT(tokenString, "wrong-secret-key")
	if err == nil {
		t.Error("expected error for wrong secret, got nil")
	}
}
