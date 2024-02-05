package openapi

import (
	_ "embed"
	"os"
	"testing"
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
	doc.GetRoute("/path/v1", "GET").
		AddResponse(
			Response{Status: 200}.WithExample(tStruct{
				Name: "apple", Int: 10,
			})).
		AddResponse(
			Response{Status: 400}.WithJSONString(`{"error":"invalid request"}`))

	route := doc.GetRoute("/path/v1", "PUT")
	route.AddRequest(
		RequestBody{}.WithJSONString(`{"account":"apple.com", "amount":10}`),
	)
	route.AddResponse(Response{Status: 200})

	doc.JSON()
}
