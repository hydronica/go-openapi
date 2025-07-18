package openapi

// Security scheme types
const (
	SecurityTypeAPIKey = "apiKey"
	SecurityTypeHTTP   = "http"
	SecurityTypeOAuth2 = "oauth2"
	SecurityTypeOpenID = "openIdConnect"
)

// HTTP authentication schemes
const (
	HTTPSchemeBearer = "bearer"
	HTTPSchemeBasic  = "basic"
)

// API key locations
const (
	APIKeyInQuery  = "query"
	APIKeyInHeader = "header"
	APIKeyInCookie = "cookie"
)

// Common bearer token formats
const (
	BearerFormatJWT = "JWT"
)

func (o *OpenAPI) AddSecurity(security string) {
	if o.Security == nil {
		o.Security = make([]SecurityRequirement, 0)
	}

	// Create a security requirement map with empty scopes for the given security scheme
	requirement := SecurityRequirement{
		security: {},
	}

	o.Security = append(o.Security, requirement)
}

// AddSecurityScheme adds a security scheme to the OpenAPI specification
func (o *OpenAPI) AddSecurityScheme(name string, scheme SecurityScheme) {
	// if the security schemes are not defined, create them
	if o.Components.SecuritySchemes == nil {
		o.Components.SecuritySchemes = make(map[string]SecurityScheme)
	}
	o.Components.SecuritySchemes[name] = scheme
}

// AddAPIKeyAuth adds an API key authentication scheme
// name is your unique identifier for the security scheme
// keyName is the name of the API key
// location is the location of the API key i.e., APIKeyInQuery, APIKeyInHeader, or APIKeyInCookie
// description is the description of the security scheme
func (o *OpenAPI) AddAPIKeyAuth(name, keyName, location, description string) {
	scheme := SecurityScheme{
		Type:        SecurityTypeAPIKey,
		Name:        keyName,
		In:          location, // APIKeyInQuery, APIKeyInHeader, or APIKeyInCookie
		Description: description,
	}
	o.AddSecurityScheme(name, scheme)
}

// AddBearerAuth adds a bearer token authentication scheme
// name is your unique identifier for the security scheme
// bearerFormat is the format of the bearer token i.e., BearerFormatJWT
// description is the description of the security scheme
func (o *OpenAPI) AddBearerAuth(name, bearerFormat, description string) {
	scheme := SecurityScheme{
		Type:         SecurityTypeHTTP,
		Scheme:       HTTPSchemeBearer,
		BearerFormat: bearerFormat,
		In:           APIKeyInHeader,
		Description:  description,
	}
	o.AddSecurityScheme(name, scheme)
}

// AddBasicAuth adds a basic authentication scheme
func (o *OpenAPI) AddBasicAuth(name string, description string) {
	scheme := SecurityScheme{
		Type:        SecurityTypeHTTP,
		Scheme:      HTTPSchemeBasic,
		Description: description,
	}
	o.AddSecurityScheme(name, scheme)
}

// AddOAuth2Auth adds an OAuth2 authentication scheme
func (o *OpenAPI) AddOAuth2Auth(name string, flows *Flows, description string) {
	scheme := SecurityScheme{
		Type:        SecurityTypeOAuth2,
		Flows:       flows,
		Description: description,
	}
	o.AddSecurityScheme(name, scheme)
}

// AddOpenIDConnectAuth adds an OpenID Connect authentication scheme
func (o *OpenAPI) AddOpenIDConnectAuth(name, openIDConnectURL, description string) {
	scheme := SecurityScheme{
		Type:             SecurityTypeOpenID,
		OpenIDConnectURL: openIDConnectURL,
		Description:      description,
	}
	o.AddSecurityScheme(name, scheme)
}

// AddSecurityRequirement adds a security requirement to the OpenAPI specification
// For non-OAuth2 schemes, pass an empty slice for scopes
// For OAuth2 schemes, pass the required scopes
func (o *OpenAPI) AddSecurityRequirement(schemeName string, scopes []string) {
	if o.Security == nil {
		o.Security = make([]SecurityRequirement, 0)
	}

	requirement := SecurityRequirement{
		schemeName: scopes,
	}

	o.Security = append(o.Security, requirement)
}

// AddMultipleSecurityRequirement adds a security requirement with multiple schemes (AND logic)
// All schemes in the map must be satisfied
func (o *OpenAPI) AddMultipleSecurityRequirement(schemes map[string][]string) {
	if o.Security == nil {
		o.Security = make([]SecurityRequirement, 0)
	}

	o.Security = append(o.Security, SecurityRequirement(schemes))
}
