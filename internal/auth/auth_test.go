package auth_test

import (
	"testing"

	"github.com/aleksaelezovic/chirpy/internal/auth"
)

func TestPasswordHashing(t *testing.T) {
	plainPassword := "password123"
	hashedPassword, err := auth.HashPassword(plainPassword)
	if err != nil {
		t.Errorf("Error hashing password: %v", err)
	}
	if hashedPassword == "" {
		t.Errorf("Hashed password is empty")
	}
	ok, err := auth.VerifyPassword(plainPassword, hashedPassword)
	if err != nil {
		t.Errorf("Error verifying password: %v", err)
	}
	if !ok {
		t.Errorf("Password verification failed")
	}
}
