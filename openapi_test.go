package openapi

import (
	_ "embed"
	"encoding/json"
	"os"
	"testing"

	"github.com/hydronica/trial"
)

//go:embed swagger.example.json
var jsonfile string

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

func ExampleNew() {
	type tStruct struct {
		Name string `json:"name"`
		Int  int    `json:"count"`
	}

	doc := New("doc", "1.0.0", "about me")
	doc.GetRoute("/path/v1", "get").
		AddResponse(
			Response{Status: 200}.WithExample(tStruct{
				Name: "apple", Int: 10,
			})).
		AddResponse(
			Response{Status: 400}.WithJSONString(`{"error":"invalid request"}`))

	route := doc.GetRoute("/path/v1", "put")
	route.AddRequest(
		RequestBody{}.WithJSONString(`{"account":"apple.com", "amount":10}`),
	)
	route.AddResponse(Response{Status: 200})

	doc.JSON()
}

type myStruct struct {
	Name    string `json:"name" desc:"individual name"`
	Age     int    `json:"age" desc:"age in earth years"`
	Country string `json:"country" desc:"3 character ISO Code"`
}

func TestRequestBodySchema(t *testing.T) {
	fn := func(in RequestBody) (string, error) {
		b, err := json.Marshal(in.Content["application/json"].Schema)
		return string(b), err
	}
	cases := trial.Cases[RequestBody, string]{
		"JSONString": {
			Input:    RequestBody{}.WithJSONString(`{"name":"bob","age":99,"country":"USA"}`),
			Expected: `{"title":"fd4b3d4f5cce2e6d","type":"object","properties":{"age":{"type":"number"},"country":{"type":"string"},"name":{"type":"string"}}}`,
		},
		"struct": {
			Input:    RequestBody{}.WithExample(myStruct{}),
			Expected: `{"title":"openapi.myStruct","type":"object","properties":{"age":{"type":"integer","description":"age in earth years"},"country":{"type":"string","description":"3 character ISO Code"},"name":{"type":"string","description":"individual name"}}}`,
		},
		"mapEx": {
			Input: RequestBody{}.WithExample(map[string]Example{
				"age":     {Value: 12, Desc: "age in earth years"},
				"country": {Value: "USA", Desc: "3 character ISO Code"},
				"name":    {Value: "bob", Desc: "individual name"},
			}),
			Expected: `{"title":"fd4b3d4f5cce2e6d","type":"object","properties":{"age":{"type":"integer","description":"age in earth years"},"country":{"type":"string","description":"3 character ISO Code"},"name":{"type":"string","description":"individual name"}}}`,
		},
		"map*Ex": {
			Input: RequestBody{}.WithExample(map[string]*Example{
				"age":     {Value: 12, Desc: "age in earth years"},
				"country": {Value: "USA", Desc: "3 character ISO Code"},
				"name":    {Value: "bob", Desc: "individual name"},
			}),
			Expected: `{"title":"fd4b3d4f5cce2e6d","type":"object","properties":{"age":{"type":"integer","description":"age in earth years"},"country":{"type":"string","description":"3 character ISO Code"},"name":{"type":"string","description":"individual name"}}}`,
		},
	}
	trial.New(fn, cases).SubTest(t)
}
