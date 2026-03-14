package github

import (
	"fmt"
	"net/http"
	"strings"
)

type authorizationTransport struct {
	Token     string
	UserAgent string
}

func (a *authorizationTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "token "+a.Token)

	userAgent := strings.TrimSpace(a.UserAgent)
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	} else {
		req.Header.Set("User-Agent", "go-kweb-lang")
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("github transport round trip failed: %w", err)
	}

	return resp, nil
}
