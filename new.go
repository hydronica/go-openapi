package openapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
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

func NewFromJson2(spec string) (api *OpenAPI2, err error) {
	api = &OpenAPI2{
		Routes: make(router),
	}
	err = json.Unmarshal([]byte(spec), &api)
	if err != nil {
		return nil, fmt.Errorf("error with unmarshal %w", err)
	}
	return api, nil
}

// OpenAPI2 represents the definition of the openapi specification 3.0.3
type OpenAPI2 struct {
	Version string   `json:"openapi"`           // the  semantic version number of the OpenAPI Specification version
	Servers []Server `json:"servers,omitempty"` // Array of Server Objects, which provide connectivity information to a target server.
	Info    Info     `json:"info"`              // REQUIRED. Provides metadata about the API. The metadata MAY be used by tooling as required.
	Tags    []Tag    `json:"tags,omitempty"`    // A list of tags used by the specification with additional metadata
	Routes  router   `json:"paths"`             // key= path|method
	//Components   Components    `json:"components,omitempty"`   // reuseable components not used here
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"` //Additional external documentation.
}

type Schema struct {
	Title string `json:"title,omitempty"`
	Type  Type   `json:"type,omitempty"`
	//Format string `json:"format,omitempty"`
	Desc string `json:"description,omitempty"`

	// Enum []string
	// Default any
	// Pattern string
	// Example any
	Items *Schema `json:"items,omitempty"`
	Ref   string  `json:"$ref,omitempty"` // link to object, #/components/schemas/{object}

	// Property definitions MUST be a Schema Object and not a standard JSON Schema (inline or referenced).
	Properties map[string]Schema `json:"properties,omitempty"`
}

type router map[string]*Route

func (r router) MarshalJSON() ([]byte, error) {
	data := make(map[string]map[string]*Route)
	for k, v := range r {
		s := strings.Split(k, "|")
		method := map[string]*Route{s[1]: v}
		data[s[0]] = method
	}

	return json.Marshal(data)
}

func (r router) UnmarshalJSON(b []byte) error {
	data := make(map[string]map[string]*Route)
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	for k1, v := range data {
		for k2, rt := range v {

			key := k1 + "|" + k2
			rt.path = k1
			rt.method = k2
			r[key] = rt
		}
	}
	return nil
}

func (o *OpenAPI2) JSON() string {
	b, err := json.MarshalIndent(o, "", "    ")
	if err != nil {
		log.Println(err)
	}
	return string(b)
}

func (o *OpenAPI2) AddRoute(r *Route) error {
	if r.path == "" || r.method == "" {
		return errors.New("path or method cannot be empty")
	}
	key := r.path + "|" + r.method
	if _, found := o.Routes[key]; found {
		return errors.New("Route already found use GetRoute to make changes")
	}

	o.Routes[key] = r
	return nil
}

// Route is a simplified definition for managing routes in code
type Route struct {
	// internal reference
	path   string
	method string

	Tag       []string          `json:"tags,omitempty"`
	Summary   string            `json:"summary,omitempty"`
	Responses map[Code]Response `json:"responses,omitempty"` // [status_code]Response
	//Params    map[string]RouteParam // key reference for params
	Requests *RequestBody `json:"requests,omitempty"` // key reference for requests

	/* NOT CURRENTLY SUPPORT VALUES
	// operationId is an optional unique string used to identify an operation
	OperationID string  json:"operationId,omitempty"`
	//A detailed description of the operation. Use markdown for rich text representation
	Desc         string        `json:"description,omitempty"`

	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"`
	*/
}

func (r *Route) WithDetails(tag, summary string) *Route {
	r.Tag = []string{tag}
	r.Summary = summary
	return r
}

// GetRoute associated with the path and method.
// create a new Route if Route was not found.
func (o *OpenAPI2) GetRoute(path, method string) *Route {
	key := path + "|" + method
	r, found := o.Routes[key]
	if !found {
		r = &Route{path: path, method: method}
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
			Content:  Content{"invalid/json": {Examples: map[string]Example{"invalid": {Value: s}}}},
		}
	}

	return r.WithStruct(m)
}

// WithStruct takes a struct and adds a json Content to the BodyObject
func (r Response) WithStruct(i any) Response {
	m := r.Content[Json]
	m.AddExample(i)
	return r
}

// AddExample will add an example
func (m *Media) AddExample(i any) {
	if m.Examples == nil {
		m.Examples = make(map[string]Example)
	}
	schema := buildSchema(i)
	if m.Schema.Title == "" {
		m.Schema = schema
	}
	exName := schema.Title
	ex := Example{
		Desc:  schema.Desc,
		Value: i,
	}

	// create unique name if key already exists
	if _, found := m.Examples[exName]; found {
		exName = exName + strconv.Itoa(len(m.Examples))
	}

	m.Examples[exName] = ex
}

func (r *Route) AddResponse(resp Response) *Route {
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
			Content: Content{"invalid/json": {Examples: map[string]Example{"invalid": {Value: s}}}},
		}
	}

	return r.WithStruct(m)
}

func (r RequestBody) WithStruct(i any) RequestBody {
	m := r.Content[Json]
	m.AddExample(i)
	return r
}

func (r *Route) AddRequest(req RequestBody) *Route {
	r.Requests = &req
	return r
}

func (r *Route) AddQueryParam() *Route {
	return r
}

func (r *Route) AddHeaderParam() *Route {
	return r
}

func (r *Route) AddPathParam() *Route {
	return r
}

func (r *Route) AddCookieParam() *Route {
	return r
}
