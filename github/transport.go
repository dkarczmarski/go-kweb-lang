package github

import (
	"net/http"
	"strings"
)

type authorizationTransport struct {
	Token     string
	UserAgent string
}

func (a *authorizationTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "token "+a.Token)
	if len(strings.TrimSpace(a.UserAgent)) > 0 {
		req.Header.Set("User-Agent", strings.TrimSpace(a.UserAgent))
	} else {
		req.Header.Set("User-Agent", "go-kweb-lang")
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	return http.DefaultTransport.RoundTrip(req)
}
