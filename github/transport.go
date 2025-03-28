package github

import "net/http"

type authorizationTransport struct {
	Token string
}

func (a *authorizationTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "token "+a.Token)
	return http.DefaultTransport.RoundTrip(req)
}
