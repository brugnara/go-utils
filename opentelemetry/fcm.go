package opentelemetry

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// FCMTransport defines a httpTransport with a custom RoundTripper
type FCMTransport struct {
	T http.RoundTripper
	CredentialsLocation string
}

// tokenProvider contains an oauth2.TokenSource used to get a valid fcm token
type tokenProvider struct {
	tokenSource oauth2.TokenSource
}

// CustomFCMTransport returns a FCMTransport
func CustomFCMTransport(T http.RoundTripper, credentialsLocation string) *FCMTransport {
	if T == nil {
		T = http.DefaultTransport
	}
	return &FCMTransport{T, credentialsLocation}
}

// RoundTrip defines a custom round trip for FCMTransport that asks for a valid FCM token
func (adt *FCMTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	tp, err := newFCMTokenProvider(adt.CredentialsLocation)
	if err != nil {
		return nil, err
	}
	token, err := tp.token()
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer " + token)
	return adt.T.RoundTrip(req)
}

// newFCMTokenProvider gets a valid token for FCM
func newFCMTokenProvider(credentialsLocation string) (*tokenProvider, error) {
	jsonKey, err := ioutil.ReadFile(credentialsLocation)
	if err != nil {
		return nil, errors.New("fcm: failed to read credentials file at: " + credentialsLocation)
	}
	cfg, err := google.JWTConfigFromJSON(jsonKey, fcmScope)
	if err != nil {
		return nil, errors.New("fcm: failed to get JWT config for the firebase.messaging scope")
	}
	ts := cfg.TokenSource(context.Background())
	return &tokenProvider{
		tokenSource: ts,
	}, nil
}

// token gets a valid FCMToken given a certain tokenProvider
func (src *tokenProvider) token() (string, error) {
	token, err := src.tokenSource.Token()
	if err != nil {
		return "", errors.New("fcm: failed to generate Bearer token")
	}
	return token.AccessToken, nil
}