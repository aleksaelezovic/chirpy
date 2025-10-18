package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/aleksaelezovic/chirpy/internal/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

func TestPasswordVerificationFail(t *testing.T) {
	plainPassword := "password123"
	hashedPassword, err := auth.HashPassword(plainPassword)
	if err != nil {
		t.Errorf("Error hashing password: %v", err)
	}
	if hashedPassword == "" {
		t.Errorf("Hashed password is empty")
	}
	ok, err := auth.VerifyPassword("password1234", hashedPassword)
	if ok {
		t.Errorf("Expected password verification to fail")
	}
}

func TestJWTGeneration(t *testing.T) {
	userID := uuid.New()
	secret := "my-super-secret-key"
	tokenString, err := auth.MakeJWT(userID, secret, 1*time.Hour)
	if err != nil {
		t.Errorf("Error generating JWT token: %v", err)
	}
	if tokenString == "" {
		t.Errorf("Generated token is empty")
	}
	tokenUserID, err := auth.ValidateJWT(tokenString, secret)
	if err != nil {
		t.Errorf("Error validating JWT token: %v", err)
	}
	if tokenUserID != userID {
		t.Errorf("Token user ID does not match expected user ID")
	}
}

func TestJWTExpiration(t *testing.T) {
	userID := uuid.New()
	secret := "my-super-secret-key"
	tokenString, err := auth.MakeJWT(userID, secret, 1*time.Second)
	if err != nil {
		t.Errorf("Error generating JWT token: %v", err)
	}
	if tokenString == "" {
		t.Errorf("Generated token is empty")
	}
	time.Sleep(2 * time.Second)
	tokenUserID, err := auth.ValidateJWT(tokenString, secret)
	if !errors.Is(err, jwt.ErrTokenExpired) {
		t.Errorf("Expected token to be expired")
	}
	if tokenUserID != uuid.Nil {
		t.Errorf("Token user ID should be empty")
	}
}

func TestJWTInvalidSignature(t *testing.T) {
	userID := uuid.New()
	secret := "my-super-secret-key"
	tokenString, err := auth.MakeJWT(userID, secret, 1*time.Hour)
	if err != nil {
		t.Errorf("Error generating JWT token: %v", err)
	}
	if tokenString == "" {
		t.Errorf("Generated token is empty")
	}
	tokenUserID, err := auth.ValidateJWT(tokenString, "wrong-secret")
	if !errors.Is(err, jwt.ErrSignatureInvalid) {
		t.Errorf("Expected token to have invalid signature")
	}
	if tokenUserID != uuid.Nil {
		t.Errorf("Token user ID should be empty")
	}
}

func TestJWTInvalidToken(t *testing.T) {
	tokenUserID, err := auth.ValidateJWT("invalid-token", "my-secret")
	if !errors.Is(err, jwt.ErrTokenMalformed) {
		t.Errorf("Expected token to be malformed")
	}
	if tokenUserID != uuid.Nil {
		t.Errorf("Token user ID should be empty")
	}
}
