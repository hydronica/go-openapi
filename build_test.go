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
	type PrimitiveTypes struct {
		F1 int    `json:"field_one"`
		F2 string `json:"field_two"`
	}

	type TestA struct {
		F1 string   `json:"field_one"`
		F2 []string `json:"field_two"`
		F3 int      `json:"field_three"`
	}

	type TestB struct {
		F1 TestA `json:"b_field_one"` // struct
	}

	// test with an object slice
	type TestC struct {
		F1 []TestA `json:"c_field_one"`
	}

	// test with an array type
	type TestD struct {
		F1 [2]string `json:"d_field_one"`
	}

	// test with a pointer
	type TestE struct {
		F1 *TestA `json:"e_field_one"` // pointer to struct
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

	type TestP struct {
		F1 *TestF `json:"f1_pointer_field"`
		F2 *TestF `json:"f2_pointer_field"`
	}

	fn := func(i any) (Schema, error) {
		s := buildSchema(i)
		return s, nil
	}

	//var z *TestF = nil // a nil typed pointer to test
	cases := trial.Cases[any, Schema]{
		"map_test_interface": {
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
			},
			//`{"description":"test description","properties":{"customValues":{"items":{"properties":{"adate":{"type":"string"},"avalue":{"format":"int64","type":"integer"}},"type":"object"},"type":"array"},"default":{"properties":{"monthProc":{"items":{"format":"float","type":"number"},"type":"array"},"monthTrans":{"items":{"format":"float","type":"number"},"type":"array"}},"type":"object"}},"title":"test title","type":"object"}`,
		},
		/*"map_test_simple": {
			Input: map[string]string{
				"key": "value",
			},
			Expected: `{"description":"test description","properties":{"key":{"type":"string"}},"title":"test title","type":"object"}`,
		},
		"map_test_object": {
			Input: map[string]PrimitiveTypes{
				"keyvalue": {F1: 123, F2: "string value"},
			},
			Expected: `{"description":"test description","properties":{"keyvalue":{"properties":{"field_one":{"format":"int64","type":"integer"},"field_two":{"type":"string"}},"type":"object"}},"title":"test title","type":"object"}`,
		},
		"nil_typed_pointer_test": {
			Input: TestP{
				F1: &TestF{
					F1: 321,
					F2: true,
				},
				F2: z, // testing a nil typed pointer
			},

			Expected: `{"description":"test description","properties":{"f1_pointer_field":{"properties":{"f1_int":{"format":"int64","type":"integer"},"f2_bool":{"type":"boolean"}},"type":"object"},"f2_pointer_field":{}},"title":"test title","type":"object"}`,
		},
		"time_test": {
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
			Input: TestA{
				F1: "testing a",
				F2: []string{"one", "two", "three"},
				F3: 1234,
			},
			Expected: Schema{
				Type: Object.String(),
				Properties: Properties{
					"F1": {Type: String.String()},
					"F2": {Type: "array"},
					"F3": {Type: "integer", Format: "int64"},
				},
			},
		}, /*
			"object_within_object": {
				Input: TestB{
					TestA{
						F1: "testing a",
						F2: []string{"one", "two", "three"},
						F3: 1234,
					},
				},
				Expected: `{"description":"test description","properties":{"b_field_one":{"properties":{"field_one":{"type":"string"},"field_three":{"format":"int64","type":"integer"},"field_two":{"items":{"type":"string"},"type":"array"}},"type":"object"}},"title":"test title","type":"object"}`,
			},
			"pointer_object": {
				Input: &TestB{
					TestA{
						F1: "testing a",
						F2: []string{"one", "two", "three"},
						F3: 1234,
					},
				},
				Expected: `{"description":"test description","properties":{"b_field_one":{"properties":{"field_one":{"type":"string"},"field_three":{"format":"int64","type":"integer"},"field_two":{"items":{"type":"string"},"type":"array"}},"type":"object"}},"title":"test title","type":"object"}`,
			},
			"pointer_in_object": {
				Input: &TestE{
					&TestA{
						F1: "testing a",
						F2: []string{"one", "two", "three"},
						F3: 1234,
					},
				},
				Expected: `{"description":"test description","properties":{"e_field_one":{"properties":{"field_one":{"type":"string"},"field_three":{"format":"int64","type":"integer"},"field_two":{"items":{"type":"string"},"type":"array"}},"type":"object"}},"title":"test title","type":"object"}`,
			},
			"array_of_array_objects": {
				Input: []TestC{
					{[]TestA{
						{
							F1: "testing slice 1",
							F2: []string{"nine", "eight", "seven"},
							F3: 987,
						},
						{
							F1: "testing slice 2",
							F2: []string{"three", "two", "one"},
							F3: 321,
						},
					}},
				},
				Expected: `{"description":"test description","items":{"properties":{"c_field_one":{"items":{"properties":{"field_one":{"type":"string"},"field_three":{"format":"int64","type":"integer"},"field_two":{"items":{"type":"string"},"type":"array"}},"type":"object"},"type":"array"}},"type":"object"},"title":"test title","type":"array"}`,
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
