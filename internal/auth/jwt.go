package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now().UTC()

	// Create registered Claims Structure
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		Subject:   userID.String(),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		NotBefore: jwt.NewNumericDate(now),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	// Create new token with HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token
	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", fmt.Errorf("Could not sign token: %v", err)
	}

	return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {

	// Initiliaze a struct to hold the parsed claims
	claims := &jwt.RegisteredClaims{}

	// Parse the token with claims, verify the signature using the secret key
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Ensure signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method")
		}
		// Return the secret as a byte slice
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("Could not parse token: %v", err)
	}
	// Extract the subject claim (contains user's stringified UUID)
	subject, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("Could not extract subject claim: %v", err)
	}

	// Parse the subject string into a UUID
	userID, err := uuid.Parse(subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("Could not parse user ID: %v", err)
	}

	return userID, nil
}
