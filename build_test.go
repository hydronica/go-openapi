package openapi

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/hydronica/trial"
)

func TestBuildSchema(t *testing.T) {

	type Simple struct {
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
		F1 time.Time `json:"time.time"`    // time.Time is always formated as RFC3339 (unless the example has it's own custom marshal for time.Time)
		F2 Time      `json:"openapi.time"` // custom time format can be used
	}

	type TestF struct {
		F1 int  `json:"f1_int"`
		F2 bool `json:"f2_bool"`
	}

	type TestP struct {
		F1 *TestF `json:"f1_pointer_field"`
		F2 *TestF `json:"f2_pointer_field"`
	}

	type input struct {
		i     any  // test input
		print bool // print out the json value
	}

	type MapTest struct {
		F1 map[string]Simple `json:"map_field"`
	}

	fn := func(i input) (string, error) {
		s, err := BuildSchema("test title", "test description", true, i.i)
		b, _ := json.Marshal(s)
		b, _ = JSONRemarshal(b)
		if i.print {
			fmt.Println(string(b))
		}
		return string(b), err
	}

	var z *TestF = nil // a nil typed pointer to test
	cases := trial.Cases[input, string]{
		"map test simple": {
			Input: input{
				i: map[string]string{
					"key": "value",
				},
				print: false,
			},
			Expected: `{"additionalProperties":{"type":"string"},"description":"test description","example":{"key":"value"},"title":"test title","type":"object"}`,
		},
		"map test object": {
			Input: input{
				i: map[string]Simple{
					"keyvalue": {F1: 123, F2: "string value"},
				},
				print: true,
			},
			Expected: `{"additionalProperties":{"properties":{"field_one":{"format":"int64","type":"integer"},"field_two":{"type":"string"}},"type":"object"},"description":"test description","example":{"keyvalue":{"field_one":123,"field_two":"string value"}},"title":"test title","type":"object"}`,
		},
		"nil typed pointer test": {
			Input: input{
				i: TestP{
					F1: &TestF{
						F1: 321,
						F2: true,
					},
					F2: z, // testing a nil typed pointer
				},
				print: false,
			},
			Expected: `{"description":"test description","example":{"f1_pointer_field":{"f1_int":321,"f2_bool":true},"f2_pointer_field":null},"properties":{"f1_pointer_field":{"properties":{"f1_int":{"format":"int64","type":"integer"},"f2_bool":{"type":"boolean"}},"type":"object"},"f2_pointer_field":{}},"title":"test title"}`,
		},
		"time test": {
			Input: input{
				i: TestT{
					F1: time.Date(2023, time.January, 11, 0, 0, 0, 0, time.UTC),
					F2: Time{
						Time:   time.Date(2023, time.February, 2, 0, 0, 0, 0, time.UTC),
						Format: "2006-01-02",
					},
				},
				print: false,
			},
			Expected: `{"description":"test description","example":{"openapi.time":"2023-02-02","time.time":"2023-01-11T00:00:00Z"},"properties":{"openapi.time":{"format":"2006-01-02","type":"string"},"time.time":{"format":"2006-01-02T15:04:05Z07:00","type":"string"}},"title":"test title"}`,
		},
		"simple_object_test": {
			Input: input{
				i: TestA{
					F1: "testing a",
					F2: []string{"one", "two", "three"},
					F3: 1234,
				},
				print: false,
			},
			Expected: `{"description":"test description","example":{"field_one":"testing a","field_three":1234,"field_two":["one","two","three"]},"properties":{"field_one":{"type":"string"},"field_three":{"format":"int64","type":"integer"},"field_two":{"items":{"type":"string"},"type":"array"}},"title":"test title"}`,
		},
		"object_within_object": {
			Input: input{
				i: TestB{
					TestA{
						F1: "testing a",
						F2: []string{"one", "two", "three"},
						F3: 1234,
					},
				},
				print: false,
			},
			Expected: `{"description":"test description","example":{"b_field_one":{"field_one":"testing a","field_three":1234,"field_two":["one","two","three"]}},"properties":{"b_field_one":{"properties":{"field_one":{"type":"string"},"field_three":{"format":"int64","type":"integer"},"field_two":{"items":{"type":"string"},"type":"array"}},"type":"object"}},"title":"test title"}`,
		},
		"pointer_object": {
			Input: input{
				i: &TestB{
					TestA{
						F1: "testing a",
						F2: []string{"one", "two", "three"},
						F3: 1234,
					},
				},
				print: false,
			},
			Expected: `{"description":"test description","example":{"b_field_one":{"field_one":"testing a","field_three":1234,"field_two":["one","two","three"]}},"properties":{"b_field_one":{"properties":{"field_one":{"type":"string"},"field_three":{"format":"int64","type":"integer"},"field_two":{"items":{"type":"string"},"type":"array"}},"type":"object"}},"title":"test title"}`,
		},
		"pointer_in_object": {
			Input: input{
				i: &TestE{
					&TestA{
						F1: "testing a",
						F2: []string{"one", "two", "three"},
						F3: 1234,
					},
				},
				print: false,
			},
			Expected: `{"description":"test description","example":{"e_field_one":{"field_one":"testing a","field_three":1234,"field_two":["one","two","three"]}},"properties":{"e_field_one":{"properties":{"field_one":{"type":"string"},"field_three":{"format":"int64","type":"integer"},"field_two":{"items":{"type":"string"},"type":"array"}},"type":"object"}},"title":"test title"}`,
		},
		"array_of_array_objects": {
			Input: input{
				i: []TestC{
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
				print: false,
			},
			Expected: `{"description":"test description","example":[{"c_field_one":[{"field_one":"testing slice 1","field_three":987,"field_two":["nine","eight","seven"]},{"field_one":"testing slice 2","field_three":321,"field_two":["three","two","one"]}]}],"items":{"properties":{"c_field_one":{"items":{"properties":{"field_one":{"type":"string"},"field_three":{"format":"int64","type":"integer"},"field_two":{"items":{"type":"string"},"type":"array"}},"type":"object"},"type":"array"}},"type":"object"},"title":"test title","type":"array"}`,
		},
	}

	trial.New(fn, cases).SubTest(t)
}
