package restapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/oauth2l/util"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
	"google.golang.org/api/impersonate"
)

type GCPOauthConfig struct {
	scopes            []string
	serviceAccountKey string
	audience          string
}

type AzureOauthConfig struct {
	GCPOpenIDTokenConfig *impersonate.IDTokenConfig
	Scope                string
	TenantId             string
	ClientId             string
	ClientAssertionType  string `default:"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"`
	GrantType            string `default:"client_credentials"`
}

// https://learn.microsoft.com/en-us/azure/active-directory/develop/v2-oauth2-client-creds-grant-flow#third-case-access-token-request-with-a-federated-credential
type azureFederatedCredentialAccessTokenResponse struct {
	TokenType   string        `json:"token_type"`
	ExpiresIn   time.Duration `json:"expires_in"`
	AccessToken string        `json:"access_token"`
}

var openIdScopes = regexp.MustCompile("^(openid|profile|email)$")

const scopePrefix = "https://www.googleapis.com/auth/"

// Append Google OAuth scope prefix if not provided and joins
// the slice into a whitespace-separated string.
func parseScopes(scopes []string) string {
	for i := 0; i < len(scopes); i++ {
		if !strings.Contains(scopes[i], "//") && !openIdScopes.MatchString(scopes[i]) {
			scopes[i] = scopePrefix + scopes[i]
		}
	}
	return strings.Join(scopes, " ")
}

func GetGCPOauthToken(gcpOauthConfig *GCPOauthConfig) (*oauth2.Token, error) {
	var token *oauth2.Token
	var err error
	parsedScopes := parseScopes(gcpOauthConfig.scopes)

	var settings = util.Settings{
		CredentialsJSON: gcpOauthConfig.serviceAccountKey,
		AuthType:        util.AuthTypeOAuth,
		Audience:        gcpOauthConfig.audience,
		Scope:           parsedScopes,
	}

	ctx := context.Background()
	token, err = util.FetchToken(ctx, &settings)

	if err != nil {
		return nil, err
	}

	return token, nil
}

func GetGCPOpenIdToken(openIdConfig *impersonate.IDTokenConfig) (*oauth2.Token, error) {
	ctx := context.Background()
	tokenSource, err := impersonate.IDTokenSource(ctx, *openIdConfig)

	if err != nil {
		return nil, err
	}

	return tokenSource.Token()
}

func GetAzureOauthToken(azureOauthConfig *AzureOauthConfig) (*oauth2.Token, error) {
	httpClient := &http.Client{}
	endpoint := microsoft.AzureADEndpoint(azureOauthConfig.TenantId).TokenURL

	clientAssertion, err := GetGCPOpenIdToken(azureOauthConfig.GCPOpenIDTokenConfig)

	if err != nil {
		return nil, err
	}

	// https://learn.microsoft.com/en-us/azure/active-directory/develop/v2-oauth2-client-creds-grant-flow#third-case-access-token-request-with-a-federated-credential
	data := url.Values{}
	data.Set("scope", azureOauthConfig.Scope)
	data.Set("client_id", azureOauthConfig.ClientId)
	data.Set("client_assertion", clientAssertion.AccessToken)
	data.Set("client_assertion_type", azureOauthConfig.ClientAssertionType)
	data.Set("grant_type", azureOauthConfig.GrantType)
	u, _ := url.ParseRequestURI(endpoint)
	urlStr := u.String()

	authRequest, err := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode()))

	if err != nil {
		return nil, err
	}

	authRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(authRequest)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var response azureFederatedCredentialAccessTokenResponse
	err = json.Unmarshal(responseBody, &response)

	if err != nil {
		return nil, err
	}

	return &oauth2.Token{AccessToken: response.AccessToken, TokenType: response.TokenType, Expiry: time.Now().Add(response.ExpiresIn)}, nil
}
