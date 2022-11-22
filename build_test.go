package openapi

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hydronica/trial"
)

func TestBuildSchema(t *testing.T) {

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
		F1 [2]string `json:"d_field_on"`
	}

	// test with a pointer
	type TestE struct {
		F1 *TestA `json:"e_field_one"` // pointer to struct
	}

	fn := func(i any) (string, error) {
		s, err := BuildSchema("test title", "test description", true, i)
		b, _ := json.Marshal(s)
		b, _ = JSONRemarshal(b)
		fmt.Println(string(b))
		return string(b), err
	}

	cases := trial.Cases[any, string]{
		"simple_object_test": {
			Input: TestA{
				F1: "testing a",
				F2: []string{"one", "two", "three"},
				F3: 1234,
			},
			Expected: `{"description":"test description","example":{"field_one":"testing a","field_three":1234,"field_two":["one","two","three"]},"properties":{"field_one":{"type":"string"},"field_three":{"format":"int64","type":"integer"},"field_two":{"items":{"type":"string"},"type":"array"}},"title":"test title"}`,
		},
		"object_within_object": {
			Input: TestB{
				TestA{
					F1: "testing a",
					F2: []string{"one", "two", "three"},
					F3: 1234,
				},
			},
			Expected: `{"description":"test description","example":{"b_field_one":{"field_one":"testing a","field_three":1234,"field_two":["one","two","three"]}},"properties":{"b_field_one":{"properties":{"field_one":{"type":"string"},"field_three":{"format":"int64","type":"integer"},"field_two":{"items":{"type":"string"},"type":"array"}},"type":"object"}},"title":"test title"}`,
		},
		"pointer_object": {
			Input: &TestB{
				TestA{
					F1: "testing a",
					F2: []string{"one", "two", "three"},
					F3: 1234,
				},
			},
			Expected: `{"description":"test description","example":{"b_field_one":{"field_one":"testing a","field_three":1234,"field_two":["one","two","three"]}},"properties":{"b_field_one":{"properties":{"field_one":{"type":"string"},"field_three":{"format":"int64","type":"integer"},"field_two":{"items":{"type":"string"},"type":"array"}},"type":"object"}},"title":"test title"}`,
		},
		"pointer_in_object": {
			Input: &TestE{
				&TestA{
					F1: "testing a",
					F2: []string{"one", "two", "three"},
					F3: 1234,
				},
			},
			Expected: `{"description":"test description","example":{"e_field_one":{"field_one":"testing a","field_three":1234,"field_two":["one","two","three"]}},"properties":{"e_field_one":{"properties":{"field_one":{"type":"string"},"field_three":{"format":"int64","type":"integer"},"field_two":{"items":{"type":"string"},"type":"array"}},"type":"object"}},"title":"test title"}`,
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
			Expected: `{"description":"test description","example":[{"c_field_one":[{"field_one":"testing slice 1","field_three":987,"field_two":["nine","eight","seven"]},{"field_one":"testing slice 2","field_three":321,"field_two":["three","two","one"]}]}],"items":{"properties":{"c_field_one":{"items":{"properties":{"field_one":{"type":"string"},"field_three":{"format":"int64","type":"integer"},"field_two":{"items":{"type":"string"},"type":"array"}},"type":"object"},"type":"array"}},"type":"object"},"title":"test title","type":"array"}`,
		},
	}

	trial.New(fn, cases).SubTest(t)
}
