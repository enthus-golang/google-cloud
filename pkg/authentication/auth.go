package authentication

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

// idTokenSource is an oauth2.TokenSource that wraps another.
// It takes the id_token from TokenSource and passes that on as a bearer token.
type idTokenSource struct {
	TokenSource oauth2.TokenSource
}

// Token returns a token from the wrapped TokenSource and extracts the id_token.
func (s *idTokenSource) Token() (*oauth2.Token, error) {
	token, err := s.TokenSource.Token()
	if err != nil {
		return nil, err
	}

	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("token did not contain an id_token")
	}

	return &oauth2.Token{
		AccessToken: idToken,
		TokenType:   "Bearer",
		Expiry:      token.Expiry,
	}, nil
}

// IDTokenTokenSource returns an oauth2.TokenSource that fetches an identity token.
func IDTokenTokenSource(ctx context.Context, audience string) (oauth2.TokenSource, error) {
	// First, try the idtoken package, which only works for service accounts.
	ts, err := idtoken.NewTokenSource(ctx, audience)
	if err != nil {
		if !strings.Contains(err.Error(), `idtoken: unsupported credentials type`) {
			return nil, err
		}

		// If that fails, we use our Application Default Credentials to fetch an id_token on the fly.
		gts, err := google.DefaultTokenSource(ctx)
		if err != nil {
			return nil, err
		}
		ts = oauth2.ReuseTokenSource(nil, &idTokenSource{TokenSource: gts})
	}
	return ts, nil
}

// AuthTransport is an http.RoundTripper that adds an Authorization header with an identity token.
type AuthTransport struct {
	Transport http.RoundTripper
	Audience  string
}

// RoundTrip executes a single HTTP transaction and adds the Authorization header.
func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	tokenSource, err := IDTokenTokenSource(req.Context(), t.Audience)
	if err != nil {
		return nil, fmt.Errorf("failed to create token source: %w", err)
	}

	token, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	return t.Transport.RoundTrip(req)
}
