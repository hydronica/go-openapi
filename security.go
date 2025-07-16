package openapi

func (o *OpenAPI) AddSecurity(security string) {
	if o.Security == nil {
		o.Security = make([]map[string][]string, 0)
	}

	// Create a security requirement map with empty scopes for the given security scheme
	requirement := map[string][]string{
		security: {},
	}

	o.Security = append(o.Security, requirement)
}

// AddSecurityScheme adds a security scheme to the OpenAPI specification
func (o *OpenAPI) AddSecurityScheme(name string, scheme SecurityScheme) {
	// if the global security schemes are not defined, create them
	o.AddSecurity(name)

	// if the security schemes are not defined, create them
	if o.Components.SecuritySchemes == nil {
		o.Components.SecuritySchemes = make(map[string]SecurityScheme)
	}
	o.Components.SecuritySchemes[name] = scheme
}

// AddAPIKeyAuth adds an API key authentication scheme
func (o *OpenAPI) AddAPIKeyAuth(name string, keyName string, location string, description string) {
	scheme := SecurityScheme{
		Type:        "apiKey",
		Name:        keyName,
		In:          location, // "query", "header", or "cookie"
		Description: description,
	}
	o.AddSecurityScheme(name, scheme)
}

// AddBearerAuth adds a bearer token authentication scheme
func (o *OpenAPI) AddBearerAuth(name string, bearerFormat string, description string) {
	scheme := SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: bearerFormat,
		Description:  description,
	}
	o.AddSecurityScheme(name, scheme)
}

// AddBasicAuth adds a basic authentication scheme
func (o *OpenAPI) AddBasicAuth(name string, description string) {
	scheme := SecurityScheme{
		Type:        "http",
		Scheme:      "basic",
		Description: description,
	}
	o.AddSecurityScheme(name, scheme)
}

// AddOAuth2Auth adds an OAuth2 authentication scheme
func (o *OpenAPI) AddOAuth2Auth(name string, flows *Flows, description string) {
	scheme := SecurityScheme{
		Type:        "oauth2",
		Flows:       flows,
		Description: description,
	}
	o.AddSecurityScheme(name, scheme)
}

// AddOpenIDConnectAuth adds an OpenID Connect authentication scheme
func (o *OpenAPI) AddOpenIDConnectAuth(name string, openIDConnectURL string, description string) {
	scheme := SecurityScheme{
		Type:             "openIdConnect",
		OpenIDConnectURL: openIDConnectURL,
		Description:      description,
	}
	o.AddSecurityScheme(name, scheme)
}

// AddSecurityRequirement adds a security requirement to the OpenAPI specification
// For non-OAuth2 schemes, pass an empty slice for scopes
// For OAuth2 schemes, pass the required scopes
func (o *OpenAPI) AddSecurityRequirement(schemeName string, scopes []string) {
	o.addSecurityRequirementHelper(map[string][]string{
		schemeName: scopes,
	})
}

// AddMultipleSecurityRequirement adds a security requirement with multiple schemes (AND logic)
// All schemes in the map must be satisfied
func (o *OpenAPI) AddMultipleSecurityRequirement(schemes map[string][]string) {
	if o.Security == nil {
		o.Security = make([]map[string][]string, 0)
	}

	o.Security = append(o.Security, schemes)
}
