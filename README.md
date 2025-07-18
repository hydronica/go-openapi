# GoOpenAPI

A Go SDK for building OpenAPI 3.0.3 specifications programmatically. Generate complete OpenAPI documentation from your Go code with type-safe schema generation and comprehensive security support.

[![codecov](https://codecov.io/gh/hydronica/go-openapi/graph/badge.svg?token=E3I51BL34W)](https://codecov.io/gh/hydronica/go-openapi)

## Features

- **Complete OpenAPI 3.0.3 Support**: Generate valid OpenAPI specifications
- **Type-Safe Schema Generation**: Automatic schema creation from Go structs, maps, and JSON
- **Comprehensive Security**: Support for API Key, Bearer, Basic, OAuth2, and OpenID Connect
- **Parameter Management**: Path, query, header, and cookie parameters with examples
- **Request/Response Bodies**: Multiple content types and examples
- **Route Management**: Fluent API for building routes and operations
- **Schema Compilation**: Automatic schema consolidation and reference generation

## Installation

```bash
go get github.com/hydronica/go-openapi
```

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/hydronica/go-openapi"
)

func main() {
    // Create a new OpenAPI document
    doc := openapi.New("My API", "1.0.0", "A sample API")

    // Add a route
    route := doc.GetRoute("/users/{id}", "GET")
    route.AddResponse(openapi.Response{
        Status: 200,
        Desc:   "User retrieved successfully",
    }.WithExample(map[string]any{
        "id":   123,
        "name": "John Doe",
        "email": "john@example.com",
    }))

    // Generate JSON
    fmt.Println(doc.JSON())
}
```

### From Existing JSON

```go
//go:embed openapi.json
var baseSpec string

doc, err := openapi.NewFromJson(baseSpec)
if err != nil {
    log.Fatal(err)
}
```

## Schema Generation

### 1. Go Structs (Recommended)

Use Go structs with `json` and `desc` tags for the cleanest schema generation:

```go
type User struct {
    ID       int    `json:"id" desc:"Unique user identifier"`
    Name     string `json:"name" desc:"User's full name"`
    Email    string `json:"email" desc:"User's email address"`
    Active   bool   `json:"active" desc:"Whether the user is active"`
    Created  time.Time `json:"created" desc:"Account creation timestamp"`
}

route.AddRequest(openapi.RequestBody{}.WithExample(User{}))
```

**Benefits:**
- Clear, readable schema names (`User` instead of hash)
- Field descriptions from `desc` tags
- Type-safe and refactor-friendly

### 2. JSON Strings

Convert JSON strings directly to schemas:

```go
route.AddRequest(openapi.RequestBody{}.WithJSONString(`{
    "name": "John Doe",
    "age": 30,
    "email": "john@example.com"
}`))
```

**Benefits:**
- Quick prototyping
- Easy for simple schemas

**Limitations:**
- Auto-generated hash names
- No field descriptions

### 3. Example Maps

Use `map[string]Example` for detailed field descriptions:

```go
route.AddRequest(openapi.RequestBody{}.WithExample(map[string]openapi.Example{
    "name": {Value: "John Doe", Desc: "User's full name"},
    "age":  {Value: 30, Desc: "User's age in years"},
    "email": {Value: "john@example.com", Desc: "User's email address"},
}))
```

**Benefits:**
- Field-level descriptions
- Flexible value types

**Limitations:**
- Auto-generated hash names
- More verbose syntax

## Route Management

### Creating Routes

```go
// Get or create a route
route := doc.GetRoute("/users/{id}", "GET")

// Add tags for grouping
route.Tags("users", "management")

// Add summary
route.Summary = "Get user by ID"
```

### Path Parameters

Path parameters are automatically detected from the path:

```go
route := doc.GetRoute("/users/{id}/posts/{postId}", "GET")
// Automatically creates path parameters for 'id' and 'postId'

// Add examples for path parameters
route.PathParam("id", 123, "User ID")
route.PathParam("postId", "abc-123", "Post ID")
```

### Query Parameters

```go
// Single parameter
route.QueryParam("limit", 10, "Maximum number of results")
route.QueryParam("offset", 0, "Number of results to skip")

// Multiple parameters from struct
type QueryParams struct {
    Limit  int    `json:"limit" desc:"Maximum results"`
    Offset int    `json:"offset" desc:"Results to skip"`
    Search string `json:"search" desc:"Search term"`
}
route.QueryParams(QueryParams{Limit: 10, Offset: 0, Search: "example"})

// Multiple parameters from map
route.QueryParams(map[string]any{
    "limit":  10,
    "offset": 0,
    "search": "example",
})
```

### Header Parameters

```go
route.HeaderParam("X-API-Version", "v1", "API version")
route.HeaderParam("X-Request-ID", "abc-123", "Request identifier")
```

### Cookie Parameters

```go
route.CookieParam("session", "abc123", "Session identifier")
```

### Multiple Parameter Examples

```go
// Multiple examples for a single parameter
route.QueryParam("status", []string{"active", "inactive", "pending"}, "User status")

// Using Example struct for custom names
route.QueryParam("priority", []openapi.Example{
    {Summary: "low", Value: 1},
    {Summary: "medium", Value: 5},
    {Summary: "high", Value: 10},
}, "Priority level")
```

## Request and Response Bodies

### Request Bodies

```go
// From struct
type CreateUserRequest struct {
    Name  string `json:"name" desc:"User's name"`
    Email string `json:"email" desc:"User's email"`
}

route.AddRequest(openapi.RequestBody{
    Desc: "User creation data",
    Required: true,
}.WithExample(CreateUserRequest{}))

// Multiple examples
route.AddRequest(openapi.RequestBody{}.
    WithNamedExample("admin", CreateUserRequest{Name: "Admin", Email: "admin@example.com"}).
    WithNamedExample("user", CreateUserRequest{Name: "User", Email: "user@example.com"}))
```

### Response Bodies

```go
// Success response
route.AddResponse(openapi.Response{
    Status: 200,
    Desc:   "User created successfully",
}.WithExample(User{}))

// Error response
route.AddResponse(openapi.Response{
    Status: 400,
    Desc:   "Invalid request",
}.WithJSONString(`{"error": "validation failed", "details": "name is required"}`))

// Multiple status codes
route.AddResponse(openapi.Response{Status: 200, Desc: "Success"}.WithExample(User{}))
route.AddResponse(openapi.Response{Status: 404, Desc: "Not found"}.WithJSONString(`{"error": "user not found"}`))
route.AddResponse(openapi.Response{Status: 500, Desc: "Server error"}.WithJSONString(`{"error": "internal server error"}`))
```

## Security

### API Key Authentication

```go
// Header-based API key
doc.AddAPIKeyAuth("YourUniqueNameApiKeyAuth", "X-API-Key", openapi.APIKeyInHeader, "API key for authentication")

// Query parameter API key
doc.AddAPIKeyAuth("ApiKeyQuery", "api_key", openapi.APIKeyInQuery, "API key as query parameter")

// Cookie-based API key
doc.AddAPIKeyAuth("ApiKeyCookie", "auth_token", openapi.APIKeyInCookie, "API key in cookie")
```

### Bearer Token Authentication

```go
doc.AddBearerAuth("BearerAuth", openapi.BearerFormatJWT, "Bearer token authentication")
```

### Basic Authentication

```go
doc.AddBasicAuth("BasicAuth", "HTTP Basic authentication")
```

### OAuth2 Authentication

```go
flows := &openapi.Flows{
    AuthorizationCode: &openapi.Flow{
        AuthorizationURL: "https://example.com/oauth/authorize",
        TokenURL:         "https://example.com/oauth/token",
        Scopes: map[string]string{
            "read":  "Read access",
            "write": "Write access",
            "admin": "Admin access",
        },
    },
}
doc.AddOAuth2Auth("OAuth2", flows, "OAuth2 authentication")
```

### OpenID Connect

```go
doc.AddOpenIDConnectAuth("OpenIDConnect", "https://example.com/.well-known/openid_configuration", "OpenID Connect authentication")
```

### Security Requirements

```go
// Single security requirement
doc.AddSecurityRequirement("ApiKeyAuth", []string{})

// OAuth2 with scopes
doc.AddSecurityRequirement("OAuth2", []string{"read", "write"})

// Multiple security schemes (AND logic)
doc.AddMultipleSecurityRequirement(map[string][]string{
    "ApiKeyAuth": {},
    "OAuth2":     {"read"},
})
```

### Available Constants

The library provides constants for common security values:

```go
// Security scheme types
openapi.SecurityTypeAPIKey    // "apiKey"
openapi.SecurityTypeHTTP      // "http"
openapi.SecurityTypeOAuth2    // "oauth2"
openapi.SecurityTypeOpenID    // "openIdConnect"

// HTTP authentication schemes
openapi.HTTPSchemeBearer     // "bearer"
openapi.HTTPSchemeBasic      // "basic"

// API key locations
openapi.APIKeyInQuery        // "query"
openapi.APIKeyInHeader       // "header"
openapi.APIKeyInCookie       // "cookie"

// Common bearer token formats
openapi.BearerFormatJWT // "JWT"
```

## Advanced Features

### Tags

```go
doc.AddTags(
    openapi.Tag{
        Name: "users",
        Desc: "User management operations",
    },
    openapi.Tag{
        Name: "posts",
        Desc: "Post management operations",
    },
)
```

### Servers

```go
doc.Servers = []openapi.Server{
    {
        URL:  "https://api.example.com/v1",
        Desc: "Production server",
    },
    {
        URL:  "https://staging-api.example.com/v1",
        Desc: "Staging server",
    },
}
```

### Path Conversion

Convert Go-style paths to OpenAPI format:

```go
// Convert ":id" to "{id}"
path := openapi.CleanPath("/users/:id/posts/:postId")
// Result: "/users/{id}/posts/{postId}"
```

### Custom Time Formatting

```go
type Event struct {
    Name string           `json:"name"`
    Date openapi.Time     `json:"date"`
}

event := Event{
    Name: "Meeting",
    Date: openapi.Time{Time: time.Now(), Format: "2006-01-02"},
}
```

### Schema Compilation

Compile the document to consolidate schemas and validate:

```go
if err := doc.Compile(); err != nil {
    log.Printf("Validation errors: %v", err)
}
```

### Custom Schema Names

```go
// Set custom name for JSON schema
jsonData := openapi.JSONString(`{"name": "value"}`).SetName("CustomSchema")
route.AddRequest(openapi.RequestBody{}.WithExample(jsonData))
```

## Complete Example

```go
package main

import (
    "fmt"
    "time"
    "github.com/hydronica/go-openapi"
)

type User struct {
    ID       int       `json:"id" desc:"Unique user identifier"`
    Name     string    `json:"name" desc:"User's full name"`
    Email    string    `json:"email" desc:"User's email address"`
    Active   bool      `json:"active" desc:"Whether the user is active"`
    Created  time.Time `json:"created" desc:"Account creation timestamp"`
}

type CreateUserRequest struct {
    Name  string `json:"name" desc:"User's name"`
    Email string `json:"email" desc:"User's email"`
}

type ErrorResponse struct {
    Error   string `json:"error" desc:"Error message"`
    Details string `json:"details" desc:"Error details"`
}

func main() {
    // Create document
    doc := openapi.New("User API", "1.0.0", "A simple user management API")

    // Add security
    doc.AddBearerAuth("BearerAuth", openapi.BearerFormatJWT, "Bearer token authentication")

    // Add tags
    doc.AddTags(openapi.Tag{
        Name: "users",
        Desc: "User management operations",
    })

    // GET /users
    listRoute := doc.GetRoute("/users", "GET")
    listRoute.Tags("users")
    listRoute.Summary = "List users"
    listRoute.QueryParam("limit", 10, "Maximum number of users to return")
    listRoute.QueryParam("offset", 0, "Number of users to skip")
    listRoute.AddResponse(openapi.Response{
        Status: 200,
        Desc:   "List of users",
    }.WithExample([]User{{ID: 1, Name: "John Doe", Email: "john@example.com", Active: true}}))

    // POST /users
    createRoute := doc.GetRoute("/users", "POST")
    createRoute.Tags("users")
    createRoute.Summary = "Create user"
    createRoute.AddRequest(openapi.RequestBody{
        Desc:     "User creation data",
        Required: true,
    }.WithExample(CreateUserRequest{Name: "Jane Doe", Email: "jane@example.com"}))
    createRoute.AddResponse(openapi.Response{
        Status: 201,
        Desc:   "User created successfully",
    }.WithExample(User{ID: 2, Name: "Jane Doe", Email: "jane@example.com", Active: true}))
    createRoute.AddResponse(openapi.Response{
        Status: 400,
        Desc:   "Invalid request",
    }.WithExample(ErrorResponse{Error: "validation failed", Details: "name is required"}))

    // GET /users/{id}
    getRoute := doc.GetRoute("/users/{id}", "GET")
    getRoute.Tags("users")
    getRoute.Summary = "Get user by ID"
    getRoute.PathParam("id", 123, "User ID")
    getRoute.AddResponse(openapi.Response{
        Status: 200,
        Desc:   "User details",
    }.WithExample(User{ID: 1, Name: "John Doe", Email: "john@example.com", Active: true}))
    getRoute.AddResponse(openapi.Response{
        Status: 404,
        Desc:   "User not found",
    }.WithExample(ErrorResponse{Error: "not found", Details: "user with id 123 not found"}))

    // Compile and validate
    if err := doc.Compile(); err != nil {
        fmt.Printf("Validation errors: %v\n", err)
    }

    // Output JSON
    fmt.Println(doc.JSON())
}
```

## API Reference

### Core Types

- `OpenAPI`: Main document structure
- `Route`: Individual API endpoint
- `RequestBody`: Request body definition
- `Response`: Response definition
- `Param`: Parameter definition
- `Schema`: Data type definition
- `Example`: Example value with description

### Methods

- `New(title, version, description)`: Create new document
- `NewFromJson(spec)`: Create from existing JSON
- `GetRoute(path, method)`: Get or create route
- `AddResponse(response)`: Add response to route
- `AddRequest(request)`: Add request body to route
- `WithExample(value)`: Add example to request/response
- `WithJSONString(json)`: Add JSON string example
- `JSON()`: Generate JSON output
- `Compile()`: Validate and consolidate schemas

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.