package openapi

import (
	_ "embed"
	"encoding/json"
	"github.com/hydronica/trial"
	"os"
	"testing"
	"time"
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

	// test a time type
	type TestT struct {
		F1 time.Time `json:"time.time" format:"2006-01-02"` // time.Time is always formated as RFC3339 (unless the example has it's own custom marshal for time.Time)
		F2 Time      `json:"openapi.time"`                  // custom time format can be used
	}

	type TestF struct {
		F1 int  `json:"f1_int"`
		F2 bool `json:"f2_bool"`
	}

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
				Type: "object",
				Properties: map[string]Schema{
					"customValues": {
						Type: Array,
						Items: &Schema{
							Type: Object,
							Properties: map[string]Schema{
								"adate":  {Type: "string"},
								"avalue": {Type: "integer"},
							},
						}},
					"default": {
						Type: Object,
						Properties: map[string]Schema{
							"monthTrans": {Type: Array, Items: &Schema{Type: Number}},
							"monthProc":  {Type: Array, Items: &Schema{Type: Number}},
						}},
				},
			},
		},
		"map_simple": {
			Input: map[string]string{
				"key": "value",
			},
			Expected: Schema{
				Type:       "object",
				Properties: map[string]Schema{"key": {Type: "string"}},
			},
		},
		"map_object": {
			Input: map[string]Primitives{
				"key": {},
			},
			Expected: Schema{
				Type: Object,
				Properties: map[string]Schema{
					"key": {
						Type: "object",
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
				Type: "object",
				Properties: map[string]Schema{
					"F2": {Type: "object", Properties: map[string]Schema{
						"f1_int":  {Type: "integer"},
						"f2_bool": {Type: "boolean"},
					}},
				},
			},
		},
		/*"time_test": {
			Input: TestT{
				F1: time.Date(2023, time.January, 11, 0, 0, 0, 0, time.UTC),
				F2: Time{
					Time:   time.Date(2023, time.February, 2, 0, 0, 0, 0, time.UTC),
					Format: "2006-01-02",
				},
			},
			Expected: `{"description":"test description","properties":{"openapi.time":{"format":"2006-01-02","type":"string"},"time.time":{"format":"2006-01-02","type":"string"}},"title":"test title","type":"object"}`,
		},*/
		"simple_object": {
			Input: Primitives{},
			Expected: Schema{
				Type: Object,
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
				Type: "object",
				Properties: map[string]Schema{
					"F1": {
						Type: "object",
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
				Type: "object",
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
				Type: Object,
				Properties: map[string]Schema{
					"F1": {
						Type: Object,
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

func TestBuilder(t *testing.T) {

	type tStruct struct {
		Name string `json:"name"`
		Int  int    `json:"count"`
	}

	doc := New2("doc", "1.0.0", "about me")
	doc.GetRoute("/path/v1", "GET").
		AddResponse(
			Response{Status: 200}.WithStruct(tStruct{
				Name: "apple", Int: 10,
			})).
		AddResponse(Response{Status: 400}.WithJSONString("abcdf")).AddRequest(RequestBody{Required: false}.WithStruct(tStruct{Name: "bob", Int: 1}))

	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(b))
}

//go:embed swagger.example.json
var jsonfile string

func TestNewFromJson2(t *testing.T) {
	doc, err := NewFromJson2(jsonfile)
	if err != nil {
		t.Fatal(err)
	}

	s := doc.JSON()
	f, err := os.Create("./gen2.json")
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte(s))
	f.Close()
}

func TestNewFromJson(t *testing.T) {
	doc, err := NewFromJson(jsonfile)
	if err != nil {
		t.Fatal(err)
	}

	s := doc.JSON()
	f, err := os.Create("./gen.json")
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte(s))
	f.Close()
}
