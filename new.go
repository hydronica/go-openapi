package openapi

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

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

// Route is a simplified definition for managing routes in code
type Route struct {
	// internal reference
	path   string
	method string

	Tag       []string          `json:"tags,omitempty"`
	Summary   string            `json:"summary,omitempty"`
	Responses map[Code]Response `json:"responses,omitempty"` // [status_code]Response
	Params    Params            // key reference for params. key is name of Param
	Requests  *RequestBody      `json:"requests,omitempty"` // key reference for requests

	/* NOT CURRENTLY SUPPORT VALUES
	// operationId is an optional unique string used to identify an operation
	OperationID string  json:"operationId,omitempty"`
	//A detailed description of the operation. Use markdown for rich text representation
	Desc         string        `json:"description,omitempty"`

	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"`
	*/
}

// parsePath goes through a url path and pulls out all params
// values surrounded by {}
var regexPathParam = regexp.MustCompile(`\{([^{}]+)\}`)

func parsePath(path string) []string {
	matches := regexPathParam.FindAllStringSubmatch(path, -1)
	r := make([]string, len(matches))
	for i := 0; i < len(matches); i++ {
		r[i] = matches[i][1]
	}
	return r
}

func (r *Route) WithDetails(tag, summary string) *Route {
	r.Tag = []string{tag}
	r.Summary = summary
	return r
}

// GetRoute associated with the path and method.
// create a new Route if Route was not found.
func (o *OpenAPI) GetRoute(path, method string) *Route {
	key := path + "|" + method
	r, found := o.Routes[key]
	if !found {
		r = &Route{path: path, method: method}
		o.Routes[key] = r
	}
	return r
}

// WithJSONString takes a json string object and adds a json Content to the Response
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

// WithStruct takes a struct and adds a json Content to the Response
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

// AddPathParam adds the path params to the route
// It does not validate that the name is part of the path
// or prevent duplicate paths from being added.
func (r *Route) AddPathParam(name string, value any) *Route {
	var p Param
	defer func() {
		r.Params[name] = p
	}()
	if r.Params == nil {
		r.Params = make(Params)
		for _, k := range parsePath(r.path) {
			r.Params[k] = Param{}
		}
	}
	p, found := r.Params[name]
	if !found {
		p = Param{
			In: "path", Name: name, Desc: "err: not found in path",
			Examples: make(map[string]Example),
		}
	}
	var errMsg string
	if !isPrimitive(value) {
		errMsg = "must be primitive type"
	}
	// what should the example name be?
	exName := fmt.Sprintf("%v", value)
	// Param already found, add as another example
	if p.Name != "" {
		p.Examples[exName] = Example{Value: value, Desc: errMsg}
		return r
	}
	p.Name = name
	p.In = "path"
	if errMsg != "" {
		p.Desc = errMsg
		return r
	}

	s := buildSchema(value)
	p.Schema = &s
	p.Examples = map[string]Example{exName: {Value: value}}

	return r
}

func isPrimitive(v any) bool {
	kind := reflect.ValueOf(v).Kind()
	if kind == reflect.Pointer {
		kind = reflect.ValueOf(v).Type().Elem().Kind()
	}
	switch kind {
	case reflect.Struct, reflect.Slice, reflect.Array, reflect.Map:
		return false
	default:
		return true
	}
}

func (r *Route) AddPathParams(params map[string]any) *Route {
	for k, v := range params {
		kind := reflect.ValueOf(v).Kind()

		// if the value is a slice of all the same kind (no any/interface type)
		// then go through each value and add it to the Param as an example
		if kind == reflect.Array || kind == reflect.Slice {
			sliceKind := reflect.ValueOf(v).Type().Elem().Kind()
			if sliceKind == reflect.Interface || sliceKind == reflect.Map ||
				sliceKind == reflect.Array || sliceKind == reflect.Slice ||
				sliceKind == reflect.Struct {
				r.AddPathParam(k, v)
				continue
			}
			val := reflect.ValueOf(v)
			for i := 0; i < val.Len(); i++ {
				r.AddPathParam(k, val.Index(i).Interface())
			}
			continue
		}

		r.AddPathParam(k, v)
	}
	return r
}

func (r *Route) AddCookieParam() *Route {
	return r
}
