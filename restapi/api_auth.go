package restapi

import (
	"context"
	"regexp"
	"strings"

	"github.com/google/oauth2l/util"
	"golang.org/x/oauth2"
)

type GCPOauthConfig struct {
	scopes            []string
	serviceAccountKey string
	audience          string
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
	var settings = util.Settings{
		CredentialsJSON: gcpOauthConfig.serviceAccountKey,
		AuthType:        util.AuthTypeJWT,
		Audience:        gcpOauthConfig.audience,
		Scope:           parseScopes(gcpOauthConfig.scopes),
	}

	ctx := context.Background()
	token, err := util.FetchToken(ctx, &settings)

	if err != nil {
		return nil, err
	}

	return token, nil
}
