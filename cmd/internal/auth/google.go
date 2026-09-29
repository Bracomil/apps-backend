package auth

import (
	"context"
	"fmt"

	"cloud.google.com/go/auth/credentials/idtoken"
)

type GoogleClaims struct {
	Subject string
	Email   string
	Name    string
	Picture string
}

type GoogleVerifier struct {
	clientID string
}

func NewGoogleVerifier(clientID string) *GoogleVerifier {
	return &GoogleVerifier{clientID: clientID}
}

func (v *GoogleVerifier) Verify(ctx context.Context, credential string) (*GoogleClaims, error) {
	// Verify token
	payload, err := idtoken.Validate(ctx, credential, "")
	if err != nil {
		return nil, fmt.Errorf("invalid google id_token: %w", err)
	}

	if payload.Subject == "" {
		return nil, fmt.Errorf("invalid subject in credential")
	}

	claims := &GoogleClaims{
		Subject: payload.Subject,
	}

	if email, ok := payload.Claims["email"].(string); ok {
		claims.Email = email
	}
	if name, ok := payload.Claims["name"].(string); ok {
		claims.Name = name
	}
	if picture, ok := payload.Claims["picture"].(string); ok {
		claims.Picture = picture
	}

	if claims.Subject == "" {
		return nil, fmt.Errorf("google id_token missing subject")
	}
	if claims.Email == "" {
		return nil, fmt.Errorf("google id_token missing email")
	}

	return claims, nil
}
