package spec

import (
	"testing"

	"github.com/hydronica/trial"
)

type TestBody struct {
	ID   int
	Name string
	Val1 int
	Val2 map[string]string
	Val3 []string
	Val4 []float64
}

func TestAddRespBody(t *testing.T) {
	type input struct {
		body BodyObject
		ur   UniqueRoute
	}

	fn := func(in input) (api *OpenAPI, err error) {
		api = New("title testing", "0.0.0", "testing description") // init the base openapi struct
		in.ur, err = api.AddRoute("/testing/path", "get", "mytag", "desc", "summary")
		if err != nil {
			return nil, err
		}

		err = api.AddReqBody(in.ur, in.body)
		if err != nil {
			return nil, err
		}
		return api, err
	}

	cases := trial.Cases[input, *OpenAPI]{
		"basic": {
			Input: input{
				ur: UniqueRoute{Path: "/testing/path", Method: GET},
				body: BodyObject{
					Body: TestBody{
						ID:   123,
						Name: "test name",
						Val1: 321,
						Val2: map[string]string{
							"key1": "value1",
						},
						Val3: []string{"one", "two"},
					},
					Desc:       "testing the response description",
					HttpStatus: "200",
					Title:      "test body",
				},
			},
			Expected: &OpenAPI{
				Version: "3.0.3",
				Info: Info{
					Title:   "title testing",
					Version: "0.0.0",
					Desc:    "testing description",
				},
				Paths: Paths{
					"/testing/path": OperationMap{
						"get": Operation{
							Tags:        []string{"mytag"},
							Desc:        "desc",
							Summary:     "summary",
							OperationID: "get_/testing/path",
							Responses: Responses{
								"200": Response{
									Content: map[string]Media{
										"application/json": Media{
											Schema: Schema{
												Title: "test body",
												Desc:  "testing the response description",
												Type:  Object.String(),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	trial.New(fn, cases).Test(t)
}
