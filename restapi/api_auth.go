package restapi

import (
	"regexp"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GCPOauthConfig struct {
	scopes            []string
	serviceAccountKey string
}

var openIdScopes = regexp.MustCompile("^(openid|profile|email)$")

const scopePrefix = "https://www.googleapis.com/auth/"

func parseGCPScopes(scopes []string) []string {
	parsedScopes := []string{}
	for i := 0; i < len(scopes); i++ {
		if !strings.Contains(scopes[i], "//") && !openIdScopes.MatchString(scopes[i]) {
			parsedScopes = append(parsedScopes, scopes[i])
		}
	}

	return parsedScopes
}

func GetGCPOauthReuseTokenSource(gcpOauthConfig *GCPOauthConfig) (*oauth2.TokenSource, error) {
	tokenSource, err := google.JWTAccessTokenSourceWithScope([]byte(gcpOauthConfig.serviceAccountKey), parseGCPScopes(gcpOauthConfig.scopes)...)
	reuseTokenSource := oauth2.ReuseTokenSource(nil, tokenSource)
	return &reuseTokenSource, err
}
