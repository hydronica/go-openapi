package openapi

import (
	"encoding/json"
	"errors"
	"log"
	"strings"
)

func New2(title, version, description string) *OpenAPI2 {
	return &OpenAPI2{
		Version: "3.0.3",
		Info: Info{
			Title:   title,
			Version: version,
			Desc:    description,
		},
		Tags:   make([]Tag, 0),
		Routes: make(router),
		//ExternalDocs: &ExternalDocs{},
	}
}

// OpenAPI2 represents the definition of the openapi specification 3.0.3
type OpenAPI2 struct {
	Version      string        `json:"openapi"`                // the  semantic version number of the OpenAPI Specification version
	Tags         []Tag         `json:"tags,omitempty"`         // A list of tags used by the specification with additional metadata
	Servers      []Server      `json:"servers,omitempty"`      // Array of Server Objects, which provide connectivity information to a target server.
	Routes       router        `json:"paths"`                  // key= path|method
	Info         Info          `json:"info"`                   // REQUIRED. Provides metadata about the API. The metadata MAY be used by tooling as required.
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"` //Additional external documentation.
}

type router map[string]*route

func (r router) MarshalJSON() ([]byte, error) {
	data := make(map[string]map[string]*route)
	for k, v := range r {
		s := strings.Split(k, "|")
		method := map[string]*route{s[1]: v}
		data[s[0]] = method
	}

	return json.Marshal(data)
}

func (o *OpenAPI2) JSON() string {
	b, err := json.Marshal(o)
	if err != nil {
		log.Println(err)
	}
	return string(b)
}

func (o *OpenAPI2) AddRoute(r *route) error {
	if r.path == "" || r.method == "" {
		return errors.New("path or method cannot be empty")
	}
	key := r.path + "|" + r.method
	if _, found := o.Routes[key]; found {
		return errors.New("route already found use GetRoute to make changes")
	}

	o.Routes[key] = r
	return nil
}

// Route is a simplified definition for managing routes in code
type route struct {
	// internal reference
	path   string
	method string

	Tag       string            `json:"tag,omitempty"`
	Summary   string            `json:"summary,omitempty"`
	Responses map[Code]Response `json:"responses,omitempty"` // [status_code]Response
	//Params    map[string]RouteParam // key reference for params
	Requests *RequestBody // key reference for requests
}

func (r *route) WithDetails(tag, summary string) *route {
	r.Tag = tag
	r.Summary = summary
	return r
}

// GetRoute associated with the path and method.
// create a new route if route was not found.
func (o *OpenAPI2) GetRoute(path, method string) *route {
	key := path + "|" + method
	r, found := o.Routes[key]
	if !found {
		r = &route{path: path, method: method}
		o.Routes[key] = r
	}
	return r
}

// WithJSONString takes a json string object and adds a json Content to the BodyObject
// s is unmarshalled into a map to extract the key and value pairs
// JSONStringResp || resp.JSONString(s)
func (r Response) WithJSONString(s string) Response {
	m := make(map[string]any)
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		// return a response with the error message
		return Response{
			Status:   r.Status,
			MimeType: "invalid/json",
			Desc:     err.Error(),
			Content:  Content{"invalid/json": {Example: s}},
		}
	}

	return r.WithStruct(m)
}

// WithStruct takes a struct and adds a json Content to the BodyObject
func (r Response) WithStruct(i any) Response {
	return Response{
		MimeType: Json,
		Status:   r.Status,
		Desc:     r.Desc,
		Content: Content{Json: Media{
			Schema:  buildSchema(i),
			Example: i,
		}},
	}

}

func (r *route) AddResponse(resp Response) *route {
	if r.Responses == nil {
		r.Responses = make(map[Code]Response)
	}
	r.Responses[resp.Status] = resp
	return r
}

func (r RequestBody) WithJSONString(s string) RequestBody {
	m := make(map[string]any)
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		// return a response with the error message
		return RequestBody{
			Desc:    err.Error(),
			Content: Content{"invalid/json": {Example: s}},
		}
	}

	return r.WithStruct(m)
}

func (r RequestBody) WithStruct(i any) RequestBody {
	return RequestBody{
		Desc:     r.Desc,
		Required: r.Required,
		Content: Content{
			XForm: {
				Schema:  buildSchema(i),
				Example: i,
			},
		},
	}
}

func (r *route) AddRequest(req RequestBody) *route {
	r.Requests = &req
	return r
}
