// Copyright (c) 2025, Mads Moi-Aune <mads@moiaune.dev>
//
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type AuthMode int

const (
	AuthOAuth2 AuthMode = iota
	AuthBasic
)

type OAuth2Profile struct {
	InstanceURL string         `json:"instance_url"`
	Token       *oauth2.Token  `json:"token"`
	Config      *oauth2.Config `json:"config"`
}

func LoadOAuth2Profile() (*OAuth2Profile, error) {
	configDir, _ := getConfigDir()
	path := path.Join(configDir, "profile.json")
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var profile OAuth2Profile
	if err := json.NewDecoder(f).Decode(&profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

func (a *OAuth2Profile) Save() error {
	configDir, _ := getConfigDir()
	path := path.Join(configDir, "profile.json")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(a)
}

func newOAuth2Client(ctx context.Context, profile *OAuth2Profile) (*http.Client, error) {
	if profile.Config == nil || profile.Token == nil {
		return nil, fmt.Errorf("You are not logged in. Please run 'snowy login'.")
	}

	ts := tokenRefresher(ctx, profile)

	tok, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("Your refresh token has expired. Please run 'snowy login' to reauthenticate")
	}

	profile.Token = tok
	if err := profile.Save(); err != nil {
		return nil, fmt.Errorf("failed to save refreshed token: %w", err)
	}

	return oauth2.NewClient(ctx, ts), nil
}

type tokenSourceFunc func() (*oauth2.Token, error)

func (f tokenSourceFunc) Token() (*oauth2.Token, error) { return f() }

func tokenRefresher(ctx context.Context, profile *OAuth2Profile) oauth2.TokenSource {
	base := profile.Config.TokenSource(ctx, profile.Token)

	return oauth2.ReuseTokenSource(profile.Token, tokenSourceFunc(func() (*oauth2.Token, error) {
		tok, err := base.Token()
		if err != nil {
			return nil, err
		}

		if tok.AccessToken != profile.Token.AccessToken {
			profile.Token = tok
			if err := profile.Save(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to save refreshed token: %v\n", err)
			}
		}

		return tok, nil
	}))
}

func randString(n int) string {
	// URL-safe base64 from crypto/rand
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	// base64-URL without padding
	return base64.RawURLEncoding.EncodeToString(b)[:n]
}

func codeChallengeS256(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}

func ensureCallback(redirectURL string) (net.Listener, string, error) {
	u := redirectURL
	if !strings.HasPrefix(u, "http://127.0.0.1:") && !strings.HasPrefix(u, "http://localhost:") {
		return nil, "", fmt.Errorf("for CLI flows, use a localhost redirect, got %s", redirectURL)
	}
	// Extract host:port
	parts := strings.Split(strings.TrimPrefix(u, "http://"), "/")
	hostPort := parts[0] // localhost:1914
	l, err := net.Listen("tcp", hostPort)
	if err != nil {
		return nil, "", fmt.Errorf("listen on %s failed: %w", hostPort, err)
	}
	cbBase := "http://" + hostPort
	return l, cbBase, nil
}

func authorizeInteractive(ctx context.Context, oauthCfg *oauth2.Config) (*oauth2.Token, error) {
	// Spin up local listener
	l, _, err := ensureCallback(oauthCfg.RedirectURL)
	if err != nil {
		return nil, err
	}
	defer l.Close()

	state := randString(24)
	verifier := randString(64)
	challenge := codeChallengeS256(verifier)

	authURL := oauthCfg.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	// Try to open browser, but always print the URL for copy/paste fallback
	fmt.Printf("If your browser does not open, visit this URL manually:\n\n%s\n\n", authURL)
	_ = openBrowser(authURL)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	// Minimal one-shot callback server
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("state") != state {
				errCh <- errors.New("oauth state mismatch")
				http.Error(w, "State mismatch", http.StatusBadRequest)
				return
			}
			code := r.URL.Query().Get("code")
			if code == "" {
				errCh <- errors.New("missing code")
				http.Error(w, "Missing code", http.StatusBadRequest)
				return
			}
			fmt.Fprintln(w, "Authentication complete. You can close this window.")
			codeCh <- code
		})
		srv := &http.Server{Handler: mux}
		_ = srv.Serve(l) // will exit after first request (listener closed by defer)
	}()

	// Wait for code
	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		return nil, err
	case <-time.After(5 * time.Minute):
		return nil, errors.New("timeout waiting for OAuth callback")
	}

	// Exchange with PKCE
	tok, err := oauthCfg.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", verifier))
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}
	return tok, nil
}
