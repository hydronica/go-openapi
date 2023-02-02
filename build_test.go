package openapi

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/hydronica/trial"
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

	type input struct {
		i     any  // test input
		print bool // print out the json value
	}

	fn := func(i input) (string, error) {
		s, err := buildSchema("test title", "test description", true, i.i, nil)
		b, _ := json.Marshal(s)
		b, _ = JSONRemarshal(b)
		if i.print {
			fmt.Println(string(b))
		}
		return string(b), err
	}

	var z *TestF = nil // a nil typed pointer to test
	cases := trial.Cases[input, string]{
		"map_test_interface": {
			Input: input{
				i: map[string]any{
					"customValues": []map[string]any{
						{"adate": "2023-02-01T00:00:00Z", "avalue": 1427200},
						{"bdate": "2023-01-01T00:00:00Z", "bvalue": 1496400},
					},
					"default": map[string][]float64{
						"monthTrans": {1.1, 2.2, 3.3, 4.4},
						"monthProc":  {5.5, 6.6, 7.7, 8.8},
					},
				},
				print: false,
			},
			Expected: `{"description":"test description","properties":{"customValues":{"items":{"properties":{"adate":{"type":"string"},"avalue":{"format":"int64","type":"integer"}},"type":"object"},"type":"array"},"default":{"properties":{"monthProc":{"items":{"format":"float","type":"number"},"type":"array"},"monthTrans":{"items":{"format":"float","type":"number"},"type":"array"}},"type":"object"}},"title":"test title","type":"object"}`,
		},
		"map_test_simple": {
			Input: input{
				i: map[string]string{
					"key": "value",
				},
				print: false,
			},
			Expected: `{"description":"test description","properties":{"key":{"type":"string"}},"title":"test title","type":"object"}`,
		},
		"map_test_object": {
			Input: input{
				i: map[string]PrimitiveTypes{
					"keyvalue": {F1: 123, F2: "string value"},
				},
				print: false,
			},
			Expected: `{"description":"test description","properties":{"keyvalue":{"properties":{"field_one":{"format":"int64","type":"integer"},"field_two":{"type":"string"}},"type":"object"}},"title":"test title","type":"object"}`,
		},
		"nil_typed_pointer_test": {
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
			Expected: `{"description":"test description","properties":{"f1_pointer_field":{"properties":{"f1_int":{"format":"int64","type":"integer"},"f2_bool":{"type":"boolean"}},"type":"object"},"f2_pointer_field":{}},"title":"test title","type":"object"}`,
		},
		"time_test": {
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
			Expected: `{"description":"test description","properties":{"openapi.time":{"format":"2006-01-02","type":"string"},"time.time":{"format":"2006-01-02","type":"string"}},"title":"test title","type":"object"}`,
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
			Expected: `{"description":"test description","properties":{"field_one":{"type":"string"},"field_three":{"format":"int64","type":"integer"},"field_two":{"items":{"type":"string"},"type":"array"}},"title":"test title","type":"object"}`,
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
			Expected: `{"description":"test description","properties":{"b_field_one":{"properties":{"field_one":{"type":"string"},"field_three":{"format":"int64","type":"integer"},"field_two":{"items":{"type":"string"},"type":"array"}},"type":"object"}},"title":"test title","type":"object"}`,
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
			Expected: `{"description":"test description","properties":{"b_field_one":{"properties":{"field_one":{"type":"string"},"field_three":{"format":"int64","type":"integer"},"field_two":{"items":{"type":"string"},"type":"array"}},"type":"object"}},"title":"test title","type":"object"}`,
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
			Expected: `{"description":"test description","properties":{"e_field_one":{"properties":{"field_one":{"type":"string"},"field_three":{"format":"int64","type":"integer"},"field_two":{"items":{"type":"string"},"type":"array"}},"type":"object"}},"title":"test title","type":"object"}`,
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
			Expected: `{"description":"test description","items":{"properties":{"c_field_one":{"items":{"properties":{"field_one":{"type":"string"},"field_three":{"format":"int64","type":"integer"},"field_two":{"items":{"type":"string"},"type":"array"}},"type":"object"},"type":"array"}},"type":"object"},"title":"test title","type":"array"}`,
		},
	}

	trial.New(fn, cases).SubTest(t)
}
