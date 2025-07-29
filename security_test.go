package openapi

import (
	"testing"

	"github.com/hydronica/trial"
)

func TestAddAPIKeyAuth(t *testing.T) {
	type input struct {
		name        string
		keyName     string
		location    string
		description string
	}

	fn := func(i input) (SecurityScheme, error) {
		doc := New("Test API", "1.0.0", "Test")
		doc.AddAPIKeyAuth(i.name, i.keyName, i.location, i.description)

		scheme, exists := doc.Components.SecuritySchemes[i.name]
		if !exists {
			return SecurityScheme{}, nil
		}
		return scheme, nil
	}

	cases := trial.Cases[input, SecurityScheme]{
		"header_auth": {
			Input: input{
				name:        "ApiKeyAuth",
				keyName:     "X-API-Key",
				location:    APIKeyInHeader,
				description: "API key authentication",
			},
			Expected: SecurityScheme{
				Type:        SecurityTypeAPIKey,
				Name:        "X-API-Key",
				In:          APIKeyInHeader,
				Description: "API key authentication",
			},
		},
		"query_auth": {
			Input: input{
				name:        "QueryAuth",
				keyName:     "api_key",
				location:    APIKeyInQuery,
				description: "Query parameter API key",
			},
			Expected: SecurityScheme{
				Type:        SecurityTypeAPIKey,
				Name:        "api_key",
				In:          APIKeyInQuery,
				Description: "Query parameter API key",
			},
		},
		"cookie_auth": {
			Input: input{
				name:        "CookieAuth",
				keyName:     "auth_token",
				location:    APIKeyInCookie,
				description: "Cookie-based authentication",
			},
			Expected: SecurityScheme{
				Type:        SecurityTypeAPIKey,
				Name:        "auth_token",
				In:          APIKeyInCookie,
				Description: "Cookie-based authentication",
			},
		},
	}

	trial.New(fn, cases).SubTest(t)
}

func TestAddBearerAuth(t *testing.T) {
	type input struct {
		name         string
		bearerFormat string
		description  string
	}

	fn := func(i input) (SecurityScheme, error) {
		doc := New("Test API", "1.0.0", "Test")
		doc.AddBearerAuth(i.name, i.bearerFormat, i.description)

		scheme, exists := doc.Components.SecuritySchemes[i.name]
		if !exists {
			return SecurityScheme{}, nil
		}
		return scheme, nil
	}

	cases := trial.Cases[input, SecurityScheme]{
		"jwt_bearer": {
			Input: input{
				name:         "BearerAuth",
				bearerFormat: BearerFormatJWT,
				description:  "JWT Bearer token",
			},
			Expected: SecurityScheme{
				Type:         SecurityTypeHTTP,
				Scheme:       HTTPSchemeBearer,
				BearerFormat: BearerFormatJWT,
				In:           APIKeyInHeader,
				Description:  "JWT Bearer token",
			},
		},
		"custom_bearer": {
			Input: input{
				name:         "CustomBearer",
				bearerFormat: "Custom",
				description:  "Custom bearer token",
			},
			Expected: SecurityScheme{
				Type:         SecurityTypeHTTP,
				Scheme:       HTTPSchemeBearer,
				BearerFormat: "Custom",
				In:           APIKeyInHeader,
				Description:  "Custom bearer token",
			},
		},
	}

	trial.New(fn, cases).SubTest(t)
}

func TestAddBasicAuth(t *testing.T) {
	type input struct {
		name        string
		description string
	}

	fn := func(i input) (SecurityScheme, error) {
		doc := New("Test API", "1.0.0", "Test")
		doc.AddBasicAuth(i.name, i.description)

		scheme, exists := doc.Components.SecuritySchemes[i.name]
		if !exists {
			return SecurityScheme{}, nil
		}
		return scheme, nil
	}

	cases := trial.Cases[input, SecurityScheme]{
		"basic_auth": {
			Input: input{
				name:        "BasicAuth",
				description: "HTTP Basic authentication",
			},
			Expected: SecurityScheme{
				Type:        SecurityTypeHTTP,
				Scheme:      HTTPSchemeBasic,
				Description: "HTTP Basic authentication",
			},
		},
	}

	trial.New(fn, cases).SubTest(t)
}

func TestAddOAuth2Auth(t *testing.T) {
	type input struct {
		name        string
		flows       *Flows
		description string
	}

	fn := func(i input) (SecurityScheme, error) {
		doc := New("Test API", "1.0.0", "Test")
		doc.AddOAuth2Auth(i.name, i.flows, i.description)

		scheme, exists := doc.Components.SecuritySchemes[i.name]
		if !exists {
			return SecurityScheme{}, nil
		}
		return scheme, nil
	}

	testFlows := &Flows{
		AuthorizationCode: &Flow{
			AuthorizationURL: "https://example.com/auth",
			TokenURL:         "https://example.com/token",
			Scopes: map[string]string{
				"read":  "Read access",
				"write": "Write access",
			},
		},
	}

	cases := trial.Cases[input, SecurityScheme]{
		"oauth2_auth": {
			Input: input{
				name:        "OAuth2Auth",
				flows:       testFlows,
				description: "OAuth2 authentication",
			},
			Expected: SecurityScheme{
				Type:        SecurityTypeOAuth2,
				Flows:       testFlows,
				Description: "OAuth2 authentication",
			},
		},
	}

	trial.New(fn, cases).SubTest(t)
}

func TestAddOpenIDConnectAuth(t *testing.T) {
	type input struct {
		name             string
		openIDConnectURL string
		description      string
	}

	fn := func(i input) (SecurityScheme, error) {
		doc := New("Test API", "1.0.0", "Test")
		doc.AddOpenIDConnectAuth(i.name, i.openIDConnectURL, i.description)

		scheme, exists := doc.Components.SecuritySchemes[i.name]
		if !exists {
			return SecurityScheme{}, nil
		}
		return scheme, nil
	}

	cases := trial.Cases[input, SecurityScheme]{
		"openid_connect": {
			Input: input{
				name:             "OpenIDConnect",
				openIDConnectURL: "https://example.com/.well-known/openid_configuration",
				description:      "OpenID Connect authentication",
			},
			Expected: SecurityScheme{
				Type:             SecurityTypeOpenID,
				OpenIDConnectURL: "https://example.com/.well-known/openid_configuration",
				Description:      "OpenID Connect authentication",
			},
		},
	}

	trial.New(fn, cases).SubTest(t)
}

func TestAddSecurityRequirement(t *testing.T) {
	type input struct {
		schemeName string
		scopes     []string
	}

	fn := func(i input) ([]SecurityRequirement, error) {
		doc := New("Test API", "1.0.0", "Test")
		doc.AddSecurityRequirement(i.schemeName, i.scopes...)
		return doc.Security, nil
	}

	cases := trial.Cases[input, []SecurityRequirement]{
		"api_key_requirement": {
			Input: input{
				schemeName: "ApiKeyAuth",
				scopes:     []string{},
			},
			Expected: []SecurityRequirement{
				{
					"ApiKeyAuth": {},
				},
			},
		},
		"oauth2_with_scopes": {
			Input: input{
				schemeName: "OAuth2Auth",
				scopes:     []string{"read", "write"},
			},
			Expected: []SecurityRequirement{
				{
					"OAuth2Auth": {"read", "write"},
				},
			},
		},
	}

	trial.New(fn, cases).SubTest(t)
}

func TestAddMultipleSecurityRequirement(t *testing.T) {
	type input struct {
		schemes map[string][]string
	}

	fn := func(i input) ([]SecurityRequirement, error) {
		doc := New("Test API", "1.0.0", "Test")
		doc.AddMultipleSecurityRequirement(i.schemes)
		return doc.Security, nil
	}

	cases := trial.Cases[input, []SecurityRequirement]{
		"multiple_schemes": {
			Input: input{
				schemes: map[string][]string{
					"ApiKeyAuth": {},
					"OAuth2Auth": {"read"},
				},
			},
			Expected: []SecurityRequirement{
				{
					"ApiKeyAuth": {},
					"OAuth2Auth": {"read"},
				},
			},
		},
		"single_scheme_in_map": {
			Input: input{
				schemes: map[string][]string{
					"BearerAuth": {},
				},
			},
			Expected: []SecurityRequirement{
				{
					"BearerAuth": {},
				},
			},
		},
	}

	trial.New(fn, cases).SubTest(t)
}

func TestSecurityConstants(t *testing.T) {
	// Test that constants have the correct values
	fn := func(constant string) (string, error) {
		return constant, nil
	}

	cases := trial.Cases[string, string]{
		"SecurityTypeAPIKey": {Input: SecurityTypeAPIKey, Expected: "apiKey"},
		"SecurityTypeHTTP":   {Input: SecurityTypeHTTP, Expected: "http"},
		"SecurityTypeOAuth2": {Input: SecurityTypeOAuth2, Expected: "oauth2"},
		"SecurityTypeOpenID": {Input: SecurityTypeOpenID, Expected: "openIdConnect"},
		"HTTPSchemeBearer":   {Input: HTTPSchemeBearer, Expected: "bearer"},
		"HTTPSchemeBasic":    {Input: HTTPSchemeBasic, Expected: "basic"},
		"APIKeyInQuery":      {Input: APIKeyInQuery, Expected: "query"},
		"APIKeyInHeader":     {Input: APIKeyInHeader, Expected: "header"},
		"APIKeyInCookie":     {Input: APIKeyInCookie, Expected: "cookie"},
		"BearerFormatJWT":    {Input: BearerFormatJWT, Expected: "JWT"},
	}

	trial.New(fn, cases).SubTest(t)
}

func TestSecuritySchemeInitialization(t *testing.T) {
	fn := func(authType string) (bool, error) {
		doc := New("Test API", "1.0.0", "Test")

		// Initially, Components.SecuritySchemes should be nil
		if doc.Components.SecuritySchemes != nil {
			return false, nil
		}

		// After adding any auth method, it should be initialized
		switch authType {
		case "apikey":
			doc.AddAPIKeyAuth("test", "key", APIKeyInHeader, "desc")
		case "bearer":
			doc.AddBearerAuth("test", BearerFormatJWT, "desc")
		case "basic":
			doc.AddBasicAuth("test", "desc")
		}

		return doc.Components.SecuritySchemes != nil, nil
	}

	cases := trial.Cases[string, bool]{
		"apikey_initializes": {Input: "apikey", Expected: true},
		"bearer_initializes": {Input: "bearer", Expected: true},
		"basic_initializes":  {Input: "basic", Expected: true},
	}

	trial.New(fn, cases).SubTest(t)
}
