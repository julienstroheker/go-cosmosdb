package cosmosdb

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Azure/go-autorest/autorest/adal"
)

type Authorizer interface {
	Authorize(context.Context, *http.Request, string, string)
}

type masterKeyAuthorizer struct {
	masterKey []byte
}

func (a *masterKeyAuthorizer) Authorize(ctx context.Context, req *http.Request, resourceType, resourceLink string) {
	date := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")

	h := hmac.New(sha256.New, a.masterKey)
	fmt.Fprintf(h, "%s\n%s\n%s\n%s\n\n", strings.ToLower(req.Method), resourceType, resourceLink, strings.ToLower(date))

	req.Header.Set("Authorization", url.QueryEscape(fmt.Sprintf("type=master&ver=1.0&sig=%s", base64.StdEncoding.EncodeToString(h.Sum(nil)))))
	req.Header.Set("x-ms-date", date)
}

func NewMasterKeyAuthorizer(masterKey string) (Authorizer, error) {
	b, err := base64.StdEncoding.DecodeString(masterKey)
	if err != nil {
		return nil, err
	}

	return &masterKeyAuthorizer{masterKey: b}, nil
}

type tokenAuthorizer struct {
	token string
}

func (a *tokenAuthorizer) Authorize(ctx context.Context, req *http.Request, resourceType, resourceLink string) {
	req.Header.Set("Authorization", url.QueryEscape(a.token))
}

func NewTokenAuthorizer(token string) Authorizer {
	return &tokenAuthorizer{token: token}
}

// oauthAADAuthorizer is used to generate oauth token will be used to connect to CosmosDB
type oauthAADAuthorizer struct {
	token *adal.ServicePrincipalToken
}

func (a *oauthAADAuthorizer) Authorize(ctx context.Context, req *http.Request, resourceType, resourceLink string) {
	oauthToken, err := getTokenCredential(ctx, a.token)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", url.QueryEscape(fmt.Sprintf("type=aad&ver=1.0&sig=%s", oauthToken)))

	date := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	req.Header.Set("x-ms-date", date)
}

func NewOauthAADAuthorizer(ctx context.Context, token *adal.ServicePrincipalToken) (Authorizer) {
	return &oauthAADAuthorizer{token: token}
}

// Gets a refreshed token credential to use on authorizer
func getTokenCredential(ctx context.Context, token *adal.ServicePrincipalToken) (string, error) {
	err := token.EnsureFreshWithContext(ctx)
	if err != nil {
		return "", err
	}
	oauthToken := token.OAuthToken()
	return oauthToken, nil
}
