package openapi

import (
	"testing"

	"github.com/hydronica/trial"
)

func TestAddParams(t *testing.T) {
	type input struct {
		value any
		path  string
		pType string
	}
	fn := func(in input) ([]Param, error) {
		r := &Route{
			path: in.path,
		}

		r.addParams(in.pType, in.value)

		return r.Params.List(), nil
	}
	cases := trial.Cases[input, []Param]{
		"basic": {
			Input: input{
				pType: "path",
				path:  "/{abc}/{count}/{amount}",
				value: map[string]any{
					"abc":    "hello",
					"amount": 12.76,
					"count":  12,
				},
			},
			Expected: []Param{
				{In: "path", Name: "abc", Schema: &Schema{Type: String},
					Examples: map[string]Example{"hello": {Value: "hello"}}},
				{In: "path", Name: "amount", Schema: &Schema{Type: Number},
					Examples: map[string]Example{"12.76": {Value: 12.76}}},
				{In: "path", Name: "count", Schema: &Schema{Type: Integer},
					Examples: map[string]Example{"12": {Value: 12}}},
			},
		},
		"invalid_type": {
			Input: input{
				pType: "path",
				path:  "/{myStruct}/{map}",
				value: map[string]any{
					"myStruct": struct{ Name string }{},
					"map":      map[string]int{},
				},
			},
			Expected: []Param{
				{In: "path", Name: "map", Desc: "err: invalid type map|struct"},
				{In: "path", Name: "myStruct", Desc: "err: invalid type map|struct"},
			},
		},
		"not in path": {
			Input: input{
				pType: "path",
				path:  "/path/to/api",
				value: map[string]any{
					"apple": 123,
				},
			},
			Expected: []Param{
				{In: "path", Name: "apple", Desc: "err: not found in path",
					Schema:   &Schema{Type: Integer},
					Examples: map[string]Example{"123": {Value: 123}},
				},
			},
		},
		"n examples": {
			Input: input{
				pType: "path",
				path:  "/{fruit}/",
				value: map[string]any{
					"fruit": []string{"apple", "banana", "nectarine", "peach"},
				},
			},
			Expected: []Param{
				{
					In:     "path",
					Name:   "fruit",
					Schema: &Schema{Type: String},
					Examples: map[string]Example{
						"apple":     {Value: "apple"},
						"banana":    {Value: "banana"},
						"nectarine": {Value: "nectarine"},
						"peach":     {Value: "peach"},
					},
				},
			},
		},
		"struct": {
			Input: input{
				pType: "path",
				path:  "/{env}/{fruit}/{version}",
				value: struct {
					Env      string `json:"env"`
					Fruit    string `json:"fruit"`
					Version  int    `json:"version"`
					unexport int
					SkipMe   string `json:"-"`
				}{Env: "dev", Fruit: "pineapple", Version: 12, SkipMe: "skip"},
			},
			Expected: []Param{
				{
					In:       "path",
					Name:     "env",
					Schema:   &Schema{Type: String},
					Examples: map[string]Example{"dev": {Value: "dev"}},
				},
				{
					In:       "path",
					Name:     "fruit",
					Schema:   &Schema{Type: String},
					Examples: map[string]Example{"pineapple": {Value: "pineapple"}},
				},
				{
					In:       "path",
					Name:     "version",
					Schema:   &Schema{Type: Integer},
					Examples: map[string]Example{"12": {Value: 12}},
				},
			},
		},
		"struct_w_desc": {
			Input: input{
				pType: "query",
				value: struct {
					Name string `desc:"first name"`
					ID   int    `desc:"unique identifier"`
				}{},
			},
			Expected: []Param{
				{In: "query", Name: "ID", Desc: "unique identifier", Schema: &Schema{Type: Integer}},
				{In: "query", Name: "Name", Desc: "first name", Schema: &Schema{Type: String}},
			},
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestAddParam(t *testing.T) {
	type input struct {
		pType string
		name  string
		desc  string
		value any
	}
	fn := func(in input) ([]Param, error) {
		r := &Route{}
		r.AddParam(in.pType, in.name, in.value, in.desc)
		return r.Params.List(), nil
	}
	cases := trial.Cases[input, []Param]{
		"*string": {
			Input: input{pType: "query", name: "variety", value: trial.StringP("orange")},
			Expected: []Param{
				{Name: "variety", In: "query", Schema: &Schema{Type: String},
					Examples: map[string]Example{"orange": {Value: "orange"}}},
			},
		},
		"any slice": {
			Input: input{pType: "query", name: "list", value: []any{1, "apple", 2, "banana"}},
			Expected: []Param{
				{Name: "list", Desc: "err: invalid param, slice elem must be primitive", In: "query"},
			},
		},
		"Example struct": {
			Input: input{pType: "query", name: "food",
				value: Example{Summary: "fruit", Value: "apple"}},
			Expected: []Param{
				{Name: "food", In: "query", Examples: map[string]Example{"fruit": {Value: "apple"}}},
			},
		},
		"Examples": {
			Input: input{
				pType: "query",
				name:  "id",
				value: []Example{
					{Summary: "aid", Value: 1234},
					{Summary: "bid", Value: 4444},
					{Summary: "xid", Value: 9944},
				},
			},
			Expected: []Param{
				{
					Name:   "id",
					In:     "query",
					Schema: &Schema{Type: Integer},
					Examples: map[string]Example{
						"aid": {Value: 1234},
						"bid": {Value: 4444},
						"xid": {Value: 9944},
					}},
			},
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestParsePath(t *testing.T) {
	fn := func(in string) ([]string, error) {
		return parsePath(in), nil
	}
	cases := trial.Cases[string, []string]{
		"brackets": {
			Input:    "/cars/{carId}/drivers/{driverId}",
			Expected: []string{"carId", "driverId"},
		},
	}
	trial.New(fn, cases).SubTest(t)
}

func TestCleanPath(t *testing.T) {
	fn := func(path string) (string, error) {
		return CleanPath(path), nil
	}

	cases := trial.Cases[string, string]{
		"no_params": {
			Input:    "/no/url/params",
			Expected: "/no/url/params",
		},
		"one_param": {
			Input:    "/test/:oneparam/route",
			Expected: "/test/{oneparam}/route",
		},
		"end_param": {
			Input:    "/test/me/:oneparam",
			Expected: "/test/me/{oneparam}",
		},
		"two_params": {
			Input:    "/test/params/:one/:two",
			Expected: "/test/params/{one}/{two}",
		},
		"three_params": {
			Input:    "/test/params/:one/:two/anything/:three",
			Expected: "/test/params/{one}/{two}/anything/{three}",
		},
	}

	trial.New(fn, cases).SubTest(t)
}

func TestAddResponse(t *testing.T) {
	doc := New("t", "v", "desc")
	route := doc.GetRoute("/test", "GET")
	route.AddResponse(Response{
		Status: 200,
		Desc:   "resp desc",
	}.WithJSONString(`{"status":"ok"}`).WithJSONString(`{"status":"ok"}`))

	route.AddResponse(Response{Status: 400}.
		WithExample(struct{ Error string }{Error: "invalid request"}))

	eq, diff := trial.Equal(route, &Route{
		path:    "/test",
		method:  "get",
		Tag:     nil,
		Summary: "",
		Responses: Responses{
			200: {
				Status: 200,
				Desc:   "resp desc",
				Content: Content{Json: Media{
					Schema: Schema{
						Type:       Object,
						Title:      "2c69c864087c4000",
						Properties: map[string]Schema{"status": {Type: "string"}},
					},
					Examples: map[string]Example{
						"Example":   {Value: map[string]any{"status": "ok"}},
						"Example 1": {Value: map[string]any{"status": "ok"}},
					},
				}},
			},
			400: {
				Status: 400,
				Content: Content{Json: Media{
					Schema: Schema{
						Title:      "struct { Error string }",
						Type:       "object",
						Properties: map[string]Schema{"Error": {Type: "string"}},
					},
					Examples: map[string]Example{
						"Example": {Value: struct{ Error string }{Error: "invalid request"}},
					},
				}},
			},
		},
	})
	if !eq {
		t.Log(diff)
		t.Fail()
	}

}

func TestAddRequest(t *testing.T) {
	type form struct {
		Name  string
		Value float32
		Count int
	}
	doc := New("t", "v", "desc")
	route := doc.GetRoute("/test", "GET")
	route.AddRequest(RequestBody{
		Desc: "custom Request",
	}.WithJSONString(`{"Name":"hello world"}`))
	route.AddRequest(RequestBody{}.WithExample(form{Name: "bob", Value: 12.34, Count: -10}))
	if len(route.Requests.Content) == 2 {
		t.Fatalf("expected two Requests to be added but got %v", len(route.Requests.Content))
	}
}

func TestMarshalRoute(t *testing.T) {

	fn := func(r Router) (string, error) {
		b, err := r.MarshalJSON()
		return string(b), err
	}
	cases := trial.Cases[Router, string]{
		"multi-method": {
			Input: Router{
				"my/path|get":    &Route{},
				"my/path|delete": &Route{},
				"my/path|put":    &Route{},
			},
			Expected: `{"my/path":{"delete":{},"get":{},"put":{}}}`,
		},
	}
	trial.New(fn, cases).SubTest(t)

}

func TestRouteAddSecurityRequirement(t *testing.T) {
	type input struct {
		schemeName string
		scopes     []string
	}

	fn := func(i input) ([]SecurityRequirement, error) {
		doc := New("Test API", "1.0.0", "Test")
		route := doc.GetRoute("/users", "GET")
		route.AddSecurity(i.schemeName, i.scopes...)
		return route.Security, nil
	}

	cases := trial.Cases[input, []SecurityRequirement]{
		"api_key_no_scopes": {
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
		"bearer_no_scopes": {
			Input: input{
				schemeName: "BearerAuth",
				scopes:     nil,
			},
			Expected: []SecurityRequirement{
				{
					"BearerAuth": {},
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
		"single_scope": {
			Input: input{
				schemeName: "OAuth2Auth",
				scopes:     []string{"admin"},
			},
			Expected: []SecurityRequirement{
				{
					"OAuth2Auth": {"admin"},
				},
			},
		},
	}

	trial.New(fn, cases).SubTest(t)
}

func TestRouteAddMultipleSecurityRequirement(t *testing.T) {
	type input struct {
		schemes map[string][]string
	}

	fn := func(i input) ([]SecurityRequirement, error) {
		doc := New("Test API", "1.0.0", "Test")
		route := doc.GetRoute("/admin", "POST")
		route.AddMultipleSecurity(i.schemes)
		return route.Security, nil
	}

	cases := trial.Cases[input, []SecurityRequirement]{
		"api_key_and_oauth2": {
			Input: input{
				schemes: map[string][]string{
					"ApiKeyAuth": {},
					"OAuth2Auth": {"admin"},
				},
			},
			Expected: []SecurityRequirement{
				{
					"ApiKeyAuth": {},
					"OAuth2Auth": {"admin"},
				},
			},
		},
		"multiple_schemes_various_scopes": {
			Input: input{
				schemes: map[string][]string{
					"BearerAuth": {},
					"OAuth2Auth": {"read", "write"},
					"ApiKeyAuth": {},
				},
			},
			Expected: []SecurityRequirement{
				{
					"BearerAuth": {},
					"OAuth2Auth": {"read", "write"},
					"ApiKeyAuth": {},
				},
			},
		},
	}

	trial.New(fn, cases).SubTest(t)
}

func TestRouteMultipleSecurityRequirements(t *testing.T) {
	type input struct {
		requirements []map[string][]string
	}

	fn := func(i input) ([]SecurityRequirement, error) {
		doc := New("Test API", "1.0.0", "Test")
		route := doc.GetRoute("/flexible", "GET")

		// Add multiple different security requirement options (OR logic)
		for _, req := range i.requirements {
			route.AddMultipleSecurity(req)
		}

		return route.Security, nil
	}

	cases := trial.Cases[input, []SecurityRequirement]{
		"multiple_alternatives": {
			Input: input{
				requirements: []map[string][]string{
					{"ApiKeyAuth": {}},                   // Option 1: API Key only
					{"OAuth2Auth": {"read"}},             // Option 2: OAuth2 with read scope
					{"BearerAuth": {}, "ApiKeyAuth": {}}, // Option 3: Both Bearer and API Key
				},
			},
			Expected: []SecurityRequirement{
				{"ApiKeyAuth": {}},
				{"OAuth2Auth": {"read"}},
				{"BearerAuth": {}, "ApiKeyAuth": {}},
			},
		},
	}

	trial.New(fn, cases).SubTest(t)
}

func TestRouteClearSecurity(t *testing.T) {
	fn := func(hasInitialSecurity bool) ([]SecurityRequirement, error) {
		doc := New("Test API", "1.0.0", "Test")
		route := doc.GetRoute("/test", "DELETE")

		if hasInitialSecurity {
			route.AddSecurity("ApiKeyAuth")
			route.AddSecurity("OAuth2Auth", "admin")
		}

		route.ClearSecurity()
		return route.Security, nil
	}

	cases := trial.Cases[bool, []SecurityRequirement]{
		"clear_existing_security": {
			Input:    true,
			Expected: nil,
		},
		"clear_no_security": {
			Input:    false,
			Expected: nil,
		},
	}

	trial.New(fn, cases).SubTest(t)
}

func TestRouteSecurityWithFullIntegration(t *testing.T) {
	type routeConfig struct {
		path           string
		method         string
		securityScheme string
		scopes         []string
		addGlobalAuth  bool
	}

	fn := func(config routeConfig) (map[string]any, error) {
		doc := New("Test API", "1.0.0", "Test")

		// Add security schemes to the document
		if config.addGlobalAuth {
			doc.AddBearerAuth("BearerAuth", BearerFormatJWT, "JWT Bearer token")
			doc.AddAPIKeyAuth("ApiKeyAuth", "X-API-Key", APIKeyInHeader, "API key")
			doc.AddOAuth2Auth("OAuth2Auth", &Flows{
				AuthorizationCode: &Flow{
					AuthorizationURL: "https://example.com/auth",
					TokenURL:         "https://example.com/token",
					Scopes: map[string]string{
						"read":  "Read access",
						"write": "Write access",
						"admin": "Admin access",
					},
				},
			}, "OAuth2 authentication")
		}

		// Create route with security
		route := doc.GetRoute(config.path, config.method)
		route.AddSecurity(config.securityScheme, config.scopes...)
		route.AddResponse(Response{
			Status: 200,
			Desc:   "Success",
		})

		// Return relevant data for validation
		return map[string]any{
			"hasSecuritySchemes": len(doc.Components.SecuritySchemes) > 0,
			"routeSecurity":      route.Security,
			"routeExists":        route != nil,
		}, nil
	}

	cases := trial.Cases[routeConfig, map[string]any]{
		"bearer_auth_integration": {
			Input: routeConfig{
				path:           "/protected",
				method:         "GET",
				securityScheme: "BearerAuth",
				scopes:         []string{},
				addGlobalAuth:  true,
			},
			Expected: map[string]any{
				"hasSecuritySchemes": true,
				"routeSecurity": []SecurityRequirement{
					{"BearerAuth": {}},
				},
				"routeExists": true,
			},
		},
		"oauth2_with_scopes": {
			Input: routeConfig{
				path:           "/admin/users",
				method:         "POST",
				securityScheme: "OAuth2Auth",
				scopes:         []string{"admin", "write"},
				addGlobalAuth:  true,
			},
			Expected: map[string]any{
				"hasSecuritySchemes": true,
				"routeSecurity": []SecurityRequirement{
					{"OAuth2Auth": {"admin", "write"}},
				},
				"routeExists": true,
			},
		},
	}

	trial.New(fn, cases).SubTest(t)
}

func TestRouteSecurityInheritance(t *testing.T) {
	type testData struct {
		addGlobalSecurity bool
		addRouteSecurity  bool
	}

	fn := func(data testData) (map[string]any, error) {
		doc := New("Test API", "1.0.0", "Test")

		// Add security schemes
		doc.AddAPIKeyAuth("ApiKeyAuth", "X-API-Key", APIKeyInHeader, "API key")
		doc.AddBearerAuth("BearerAuth", BearerFormatJWT, "Bearer token")

		// Add global security if requested
		if data.addGlobalSecurity {
			doc.AddSecurityRequirement("ApiKeyAuth")
		}

		// Create route
		route := doc.GetRoute("/test", "GET")

		// Add route-specific security if requested
		if data.addRouteSecurity {
			route.AddSecurity("BearerAuth")
		}

		return map[string]any{
			"globalSecurity": doc.Security,
			"routeSecurity":  route.Security,
			"bothExist":      doc.Security != nil && route.Security != nil,
		}, nil
	}

	cases := trial.Cases[testData, map[string]any]{
		"global_only": {
			Input: testData{
				addGlobalSecurity: true,
				addRouteSecurity:  false,
			},
			Expected: map[string]any{
				"globalSecurity": []SecurityRequirement{
					{"ApiKeyAuth": {}},
				},
				"routeSecurity": []SecurityRequirement(nil),
				"bothExist":     false,
			},
		},
		"route_only": {
			Input: testData{
				addGlobalSecurity: false,
				addRouteSecurity:  true,
			},
			Expected: map[string]any{
				"globalSecurity": []SecurityRequirement(nil),
				"routeSecurity": []SecurityRequirement{
					{"BearerAuth": {}},
				},
				"bothExist": false,
			},
		},
		"both_global_and_route": {
			Input: testData{
				addGlobalSecurity: true,
				addRouteSecurity:  true,
			},
			Expected: map[string]any{
				"globalSecurity": []SecurityRequirement{
					{"ApiKeyAuth": {}},
				},
				"routeSecurity": []SecurityRequirement{
					{"BearerAuth": {}},
				},
				"bothExist": true,
			},
		},
	}

	trial.New(fn, cases).SubTest(t)
}
