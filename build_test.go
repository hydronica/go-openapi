package openapi

import (
	_ "embed"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hydronica/trial"
)

func TestBuildSchema(t *testing.T) {
	type Primitives struct {
		Int    int `json:"custom_int"`
		String string
		Bool   bool
		Number float64
	}

	type TestA struct {
		F1 string
		F2 []string
		F3 int
	}

	type TestF struct {
		F1 int  `json:"f1_int"`
		F2 bool `json:"f2_bool"`
	}
	setJSON := JSONString(`{"error": "invalid"}`).SetName("error_message")

	fn := func(i any) (Schema, error) {
		return buildSchema(i), nil
	}

	cases := trial.Cases[any, Schema]{
		"map_any": {
			Input: map[string]any{
				"customValues": []map[string]any{
					{"adate": "2023-02-01T00:00:00Z", "avalue": 1427200},
					{"bdate": "2023-01-01T00:00:00Z", "bvalue": 1496400},
				},
				"default": map[string][]float64{
					"monthTrans": {1.1, 2.2, 3.3, 4.4},
					"monthProc":  {5.5, 6.6, 7.7, 8.8},
				},
			},
			Expected: Schema{
				Type:  "object",
				Title: "d048026ab7fb3f07",
				Properties: map[string]Schema{
					"customValues": {
						Type: Array,
						Items: &Schema{
							Type:  Object,
							Title: "6e1052120b0412c8",
							Properties: map[string]Schema{
								"adate":  {Type: "string"},
								"avalue": {Type: "integer"},
							},
						}},
					"default": {
						Type:  Object,
						Title: "502deb2d72175bcd",
						Properties: map[string]Schema{
							"monthTrans": {Type: Array, Items: &Schema{Type: Number}},
							"monthProc":  {Type: Array, Items: &Schema{Type: Number}},
						}},
				},
			},
		},
		"jsonString": {
			Input: JSONString(`{"key": "value"}`),
			Expected: Schema{
				Type:       "object",
				Title:      "2292dac000000000",
				Properties: map[string]Schema{"key": {Type: "string"}},
			},
		},
		"jsonString_named": {
			Input: setJSON,
			Expected: Schema{
				Type:       "object",
				Title:      "error_message",
				Properties: map[string]Schema{"error": {Type: "string"}},
			},
		},
		/*"jsonString_array": {
			Input: JSONString(`["value1", "value2"]`),
			Expected: Schema{
				Type:  "array",
				Items: &Schema{Type: "string"},
				Title: "2292dac000000000",
			},
		},*/
		"map_simple": {
			Input: map[string]string{
				"key": "value",
			},
			Expected: Schema{
				Type:       "object",
				Title:      "2292dac000000000",
				Properties: map[string]Schema{"key": {Type: "string"}},
			},
		},
		"map_object": {
			Input: map[string]Primitives{
				"key": {},
			},
			Expected: Schema{
				Type:  Object,
				Title: "2292dac000000000", // generated hash
				Properties: map[string]Schema{
					"key": {
						Type:  "object",
						Title: "openapi.Primitives",
						Properties: map[string]Schema{
							"custom_int": {Type: Integer},
							"String":     {Type: String},
							"Bool":       {Type: Boolean},
							"Number":     {Type: Number},
						},
					},
				},
			},
		},
		"map_nil_struct": {
			Input: struct{ F2 *TestF }{
				F2: (*TestF)(nil), // testing a nil typed pointer
			},

			Expected: Schema{
				Type:  "object",
				Title: "struct { F2 *openapi.TestF }",
				Properties: map[string]Schema{
					"F2": {
						Type:  "object",
						Title: "openapi.TestF",
						Properties: map[string]Schema{
							"f1_int":  {Type: "integer"},
							"f2_bool": {Type: "boolean"},
						}},
				},
			},
		},
		"time": {
			Input: trial.TimeDay("2023-01-11"),

			Expected: Schema{
				Title: "time.Time",
				Type:  "string",
			},
		},
		"simple_object": {
			Input: Primitives{},
			Expected: Schema{
				Type:  Object,
				Title: "openapi.Primitives",
				Properties: map[string]Schema{
					"custom_int": {Type: Integer},
					"String":     {Type: String},
					"Bool":       {Type: Boolean},
					"Number":     {Type: Number},
				},
			},
		},
		"object_within_object": {
			Input: struct {
				F1 TestA
			}{
				TestA{
					F1: "testing a",
					F2: []string{"one", "two", "three"},
					F3: 1234,
				},
			},
			Expected: Schema{
				Type:  "object",
				Title: "struct { F1 openapi.TestA }",
				Properties: map[string]Schema{
					"F1": {
						Type:  "object",
						Title: "openapi.TestA",
						Properties: map[string]Schema{
							"F1": {Type: "string"},
							"F2": {Type: "array", Items: &Schema{Type: "string"}},
							"F3": {Type: "integer"},
						},
					},
				},
			},
		},
		"pointer_object": {
			Input: &TestA{
				F1: "testing a",
				F2: []string{"one", "two", "three"},
				F3: 1234,
			},
			Expected: Schema{
				Type:  "object",
				Title: "openapi.TestA",
				Properties: map[string]Schema{
					"F1": {Type: "string"},
					"F2": {Type: "array", Items: &Schema{Type: "string"}},
					"F3": {Type: "integer"},
				},
			},
		},
		"pointer_in_object": {
			Input: &struct{ F1 *TestA }{
				&TestA{
					F1: "testing a",
					F2: []string{"one", "two", "three"},
					F3: 1234,
				},
			},
			Expected: Schema{
				Type:  Object,
				Title: "struct { F1 *openapi.TestA }",
				Properties: map[string]Schema{
					"F1": {
						Type:  Object,
						Title: "openapi.TestA",
						Properties: map[string]Schema{
							"F1": {Type: "string"},
							"F2": {Type: "array", Items: &Schema{Type: "string"}},
							"F3": {Type: "integer"},
						}},
				},
			},
		},
		"array_of_array_objects": {
			Input: [][]struct {
				Name string
			}{},
			Expected: Schema{
				Type: Array,
				Items: &Schema{Type: Array,
					Items: &Schema{
						Title:      "struct { Name string }",
						Type:       Object,
						Properties: map[string]Schema{"Name": {Type: String}},
					},
				},
			},
		},
		/*"any_array": {
			Input: []any{"eholo", struct{ Name string }{Name: "abc"}},
		}, */
	}

	trial.New(fn, cases).SubTest(t)
}

func TestCompile(t *testing.T) {
	type abc struct {
		Date  time.Time
		Price float64
		Count int
	}

	fn := func(r *Route) (*OpenAPI, error) {
		o := New("", "", "")
		o.Paths[r.key()] = r
		err := o.Compile()
		return o, err
	}
	cases := trial.Cases[*Route, *OpenAPI]{
		"request-object": {
			Input: (&Route{path: "test", method: "get"}).
				AddRequest(RequestBody{}.
					WithExample(
						abc{
							Date:  trial.TimeDay("2023-11-01"),
							Price: 12.34,
							Count: 1},
					)),
			Expected: &OpenAPI{
				Paths: Router{
					"test|get": &Route{
						Requests: &RequestBody{
							Content: Content{Json: {
								Schema: Schema{Ref: "#/components/schemas/openapi.abc"},
							},
							},
						},
					},
				},
				Components: Components{
					Schemas: map[string]Schema{"openapi.abc": {
						Title: "openapi.abc",
						Type:  Object,
						Properties: map[string]Schema{
							"Count": {Type: Integer},
							"Date":  {Type: String, Title: "time.Time"},
							"Price": {Type: Number},
						},
					}},
				},
			},
		},
		"response-object": {
			Input: (&Route{path: "test", method: "get"}).AddResponse(
				Response{Status: 200}.WithExample(abc{})),
			Expected: &OpenAPI{
				Paths: Router{
					"test|get": &Route{
						Responses: Responses{
							200: Response{
								Status: 200,
								Content: Content{
									Json: {
										Schema: Schema{Ref: "#/components/schemas/openapi.abc"},
									}},
							}},
					},
				},
				Components: Components{
					Schemas: map[string]Schema{"openapi.abc": {
						Title: "openapi.abc",
						Type:  Object,
						Properties: map[string]Schema{
							"Count": {Type: Integer},
							"Date":  {Type: String, Title: "time.Time"},
							"Price": {Type: Number},
						},
					}},
				},
			},
		},
		"request-error": {
			Input: (&Route{path: "test", method: "get"}).
				AddRequest(RequestBody{}.WithJSONString("invalid")),
			ExpectedErr: errors.New(`invalid json get request at test: "invalid"`),
		},
		"response-error": {
			Input: (&Route{path: "test", method: "get"}).
				AddResponse(Response{Status: 200}.WithJSONString("invalid")),
			ExpectedErr: errors.New(`invalid json get response at test: "invalid"`),
		},
		"param-error": {
			Input:       (&Route{path: "test", method: "get"}).AddParam("query", "name", abc{}, ""),
			ExpectedErr: errors.New("query param name| err"),
		},
	}
	ignoreExamples := func(_ any) cmp.Option {
		return cmpopts.IgnoreFields(Media{}, "Examples")
	}
	trial.New(fn, cases).Comparer(
		trial.EqualOpt(
			trial.IgnoreAllUnexported,
			trial.IgnoreFields("Version", "Tags"),
			ignoreExamples,
		)).SubTest(t)
}
