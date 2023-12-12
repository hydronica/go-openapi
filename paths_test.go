package openapi

import (
	"github.com/hydronica/trial"
	"testing"
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

		r.AddParams(in.pType, in.value)

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
		r.AddParam(in.pType, in.name, in.desc, in.value)
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
	}.WithJSONString(`{"status":"ok"}`))
	route.AddResponse(Response{Status: 400}.WithExample(struct{ Error string }{Error: "invalid request"}))

	eq, diff := trial.Equal(route, &Route{
		path:    "/test",
		method:  "GET",
		Tag:     nil,
		Summary: "",
		Responses: Responses{
			200: {
				Status: 200,
				Desc:   "resp desc",
				Content: Content{Json: Media{
					Schema: Schema{
						Type:       Object,
						Title:      "map[string]interface {}",
						Properties: map[string]Schema{"status": {Type: "string"}},
					},
					Examples: map[string]Example{
						"map[string]interface {}": {
							Value: map[string]any{"status": "ok"},
						},
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
						"struct { Error string }": {
							Value: struct{ Error string }{Error: "invalid request"},
						},
					},
				}},
			},
		},
	})
	if !eq {
		t.Logf(diff)
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
