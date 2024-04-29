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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/go-autorest/autorest/adal"
)

type Authorizer interface {
	Authorize(context.Context, *http.Request, string, string) error
}

type masterKeyAuthorizer struct {
	masterKey []byte
}

func (a *masterKeyAuthorizer) Authorize(ctx context.Context, req *http.Request, resourceType, resourceLink string) error {
	date := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")

	h := hmac.New(sha256.New, a.masterKey)
	fmt.Fprintf(h, "%s\n%s\n%s\n%s\n\n", strings.ToLower(req.Method), resourceType, resourceLink, strings.ToLower(date))

	req.Header.Set("Authorization", url.QueryEscape(fmt.Sprintf("type=master&ver=1.0&sig=%s", base64.StdEncoding.EncodeToString(h.Sum(nil)))))
	req.Header.Set("x-ms-date", date)

	return nil
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

func (a *tokenAuthorizer) Authorize(ctx context.Context, req *http.Request, resourceType, resourceLink string) error {
	req.Header.Set("Authorization", url.QueryEscape(a.token))

	return nil
}

func NewTokenAuthorizer(token string) Authorizer {
	return &tokenAuthorizer{token: token}
}

// oauthAADAuthorizer is used to generate oauth token will be used to connect to CosmosDB
type oauthAADAuthorizer struct {
	token *adal.ServicePrincipalToken
}

func (a *oauthAADAuthorizer) Authorize(ctx context.Context, req *http.Request, resourceType, resourceLink string) error {
	oauthToken, err := getTokenCredential(ctx, a.token)
	if err != nil {
		return fmt.Errorf("error authorizing request using OAuth AAD Authorizer: %w", err)
	}
	setAADHeaders(req, oauthToken)

	return nil
}

func NewOauthAADAuthorizer(token *adal.ServicePrincipalToken) Authorizer {
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

type oauthMsalAADAuthorizer struct {
	token               azcore.TokenCredential
	cosmosDBInstanceURI string
}

func NewOauthMsalAADAuthorizer(token azcore.TokenCredential, cosmosDBInstanceURI string) Authorizer {
	return &oauthMsalAADAuthorizer{
		token:               token,
		cosmosDBInstanceURI: cosmosDBInstanceURI,
	}
}

func (a *oauthMsalAADAuthorizer) Authorize(ctx context.Context, req *http.Request, resourceType, resourceLink string) error {
	oauthToken, err := getMsalToken(ctx, a.token, a.cosmosDBInstanceURI)
	if err != nil {
		return fmt.Errorf("error authorizing request using OAuth AAD Authorizer: %w", err)
	}
	setAADHeaders(req, oauthToken)

	return nil
}

func getMsalToken(ctx context.Context, tokenCred azcore.TokenCredential, cosmosDBInstanceURI string) (string, error) {
	scopes, err := createScopeFromEndpoint(cosmosDBInstanceURI)
	if err != nil {
		return "", fmt.Errorf("error creating scopes: %w", err)
	}
	token, err := tokenCred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: scopes,
	})

	if err != nil {
		return "", fmt.Errorf("error getting token: %w", err)
	}
	return token.Token, nil
}

func createScopeFromEndpoint(endpoint string) ([]string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	return []string{fmt.Sprintf("%s://%s/.default", u.Scheme, u.Hostname())}, nil
}

func setAADHeaders(req *http.Request, oauthToken string) {
	req.Header.Set("Authorization", url.QueryEscape(fmt.Sprintf("type=aad&ver=1.0&sig=%s", oauthToken)))

	date := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	req.Header.Set("x-ms-date", date)
}
