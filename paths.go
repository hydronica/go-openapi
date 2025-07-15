package openapi

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Router key is path|method
type Router map[string]*Route

// Route is a simplified definition for managing routes in code
type Route struct {
	// internal reference
	path   string
	method string

	Tag       []string          `json:"tags,omitempty"`
	Summary   string            `json:"summary,omitempty"`
	Responses map[Code]Response `json:"responses,omitempty"`   // [status_code]Response
	Params    Params            `json:"parameters,omitempty"`  // key reference for params. key is name of Param
	Requests  *RequestBody      `json:"requestBody,omitempty"` // key reference for requests

	/* NOT CURRENTLY SUPPORT VALUES
	// operationId is an optional unique string used to identify an operation
	OperationID string  json:"operationId,omitempty"`
	//A detailed description of the operation. Use markdown for rich text representation
	Desc         string        `json:"description,omitempty"`

	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"`
	*/
}

func (r *Route) key() string {
	return r.path + "|" + r.method
}

func (r Router) MarshalJSON() ([]byte, error) {
	data := make(map[string]map[string]*Route)
	for k, v := range r {
		s := strings.Split(k, "|")
		path, method := s[0], s[1]
		if d, found := data[path]; !found {
			data[path] = map[string]*Route{method: v}
		} else {
			d[method] = v
		}
	}

	return json.Marshal(data)
}

func (r Router) UnmarshalJSON(b []byte) error {
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

func (r *Route) Tags(tag ...string) *Route {
	r.Tag = tag
	return r
}

// CleanPath will convert of go path like :var into
// an approved openID path {var}
func CleanPath(path string) string {
	cnt := strings.Count(path, ":")
	for c := 0; c < cnt; c++ {
		i := strings.Index(path, ":")
		if i == -1 {
			break
		}
		a := path[i:]
		e := strings.Index(a, "/")
		if e == -1 {
			e = len(a)
		}
		param := a[:e]

		path = strings.Replace(path, param, "{"+param[1:]+"}", 1)
	}

	return path
}

// GetRoute associated with the path and method.
// create a new Route if Route was not found.
// method should always be lowercase
func (o *OpenAPI) GetRoute(path, method string) *Route {
	method = strings.ToLower(method)
	key := path + "|" + method
	r, found := o.Paths[key]
	if !found {
		r = &Route{
			path:   path,
			method: method,
			Params: make(Params),
		}

		// Add any path params
		for _, k := range parsePath(r.path) {
			r.Params["path|"+k] = Param{
				Name:     k,
				In:       "path",
				Examples: make(map[string]Example),
			}
		}
		o.Paths[key] = r
	}
	return r
}

func (r *Route) AddResponse(resp Response) *Route {
	if r.Responses == nil {
		r.Responses = make(map[Code]Response)
	}
	r.Responses[resp.Status] = resp
	return r
}

// Responses for the expected responses of an operation, maps a HTTP response code to the expected response.
type Responses map[Code]Response

// Response describes a single response from an API Operation
type Response struct {
	Status Code `json:"-"`
	//MimeType MIMEType `json:"-"`

	Desc    string  `json:"description"`       // Required A short description of the response. CommonMark syntax MAY be used for rich text representation.
	Content Content `json:"content,omitempty"` // A map containing descriptions of potential response payloads. The key is a media type or media type range and the value describes it.
}

// WithJSONString takes a json string object and adds a json Content to the Response
// s is unmarshalled into a map to extract the key and value pairs
// JSONStringResp || resp.JSONString(s)
func (r Response) WithJSONString(s string) Response {
	return r.WithNamedJsonString("", s)
}

// WithExample takes a struct and adds a json Content to the Response
// name is auto generated based on the example count
func (r Response) WithExample(i any) Response {
	return r.WithNamedExample("", i)
}

// WithNamedJsonString takes a json string object and adds a json Content to the Response
// s is unmarshalled into a map to extract the key and value pairs
// JSONStringResp || resp.JSONString(s)
func (r Response) WithNamedJsonString(name string, s string) Response {
	var m any
	if s[0] == '[' && s[len(s)-1] == ']' {
		m = make([]any, 0)
	} else {
		m = make(map[string]any)
	}
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		// return a response with the error message
		return Response{
			Status:  r.Status,
			Desc:    err.Error(),
			Content: Content{"invalid/json": {Examples: map[string]Example{"invalid": {Value: s}}}},
		}
	}
	return r.WithNamedExample(name, m)
}

func (r Response) WithNamedExample(name string, i any) Response {
	if r.Content == nil {
		r.Content = make(Content)
	}
	m := r.Content[Json]
	m.AddExample(name, i)
	r.Content[Json] = m
	return r
}

// AddExample will add an example object by
// creating a schema based on the object i passed in.
// The Example name will be the title of the Schema if not provided
// and any description from added to the example as well.
func (m *Media) AddExample(name string, i any) {
	if m.Examples == nil {
		m.Examples = make(map[string]Example)
	}
	schema := buildSchema(i)
	if m.Schema.Title == "" {
		m.Schema = schema
	}
	if name == "" {
		name = "Example"
	}
	ex := Example{
		Desc:  schema.Desc,
		Value: i,
	}

	// create unique name if key already exists
	if _, found := m.Examples[name]; found {
		name = name + " " + strconv.Itoa(len(m.Examples))
	}

	m.Examples[name] = ex
}

// RequestBody describes a single request body.
type RequestBody struct {
	Desc     string  `json:"description,omitempty"` // A brief description of the request body. This could contain examples of use. CommonMark syntax MAY be used for rich text representation.
	Content  Content `json:"content"`               // REQUIRED. The content of the request body. The key is a media type or media type range and the value describes it. For requests that match multiple keys, only the most specific key is applicable. e.g. text/plain overrides text/*
	Required bool    `json:"required,omitempty"`    // Determines if the request body is required in the request. Defaults to false.
}

func (r RequestBody) WithNamedJsonString(name string, s string) RequestBody {
	return r.WithNamedExample(name, s)
}

func (r RequestBody) WithJSONString(s string) RequestBody {
	var m any
	if s[0] == '[' && s[len(s)-1] == ']' {
		m = make([]any, 0)
	} else {
		m = make(map[string]any)
	}
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		// return a response with the error message
		return RequestBody{
			Desc:    err.Error(),
			Content: Content{"invalid/json": {Examples: map[string]Example{"invalid": {Value: s}}}},
		}
	}
	return r.WithExample(m)
}

func (r RequestBody) WithExample(i any) RequestBody {
	return r.WithNamedExample("", i)
}

func (r RequestBody) WithNamedExample(name string, i any) RequestBody {
	if r.Content == nil {
		r.Content = make(Content)
	}
	m := r.Content[Json]
	m.AddExample(name, i)
	r.Content[Json] = m
	return r
}

func (r *Route) AddRequest(req RequestBody) *Route {
	r.Requests = &req
	return r
}

type ParamSetter func() Param

type Params map[string]Param

// Param see https://swagger.io/docs/specification/describing-parameters/
// - Path /user/{id}
// - Query /user?role=admin
// - header X-MyHeader: Value
// - cookie
type Param struct {
	Name string `json:"name,omitempty"`        // REQUIRED. The name of the parameter.
	Desc string `json:"description,omitempty"` // A brief description of the parameter.

	In string `json:"in"` // REQUIRED. Param Type: "query", "header", "path" or "cookie".

	Schema   *Schema            `json:"schema,omitempty"` // The schema defining the param
	Examples map[string]Example `json:"examples"`         // Examples of the parameter’s potential value.

	// NOT CURRENTLY SUPPORTED
	//Style    string             `json:"style,omitempty"`       // Describes how the parameter value will be serialized depending on the type of the parameter value. Default values (based on value of in): for query - form; for path - simple; for header - simple; for cookie - form.
	//Required bool               `json:"required"`              // Determines whether this parameter is mandatory. If the parameter location is "path", this property is REQUIRED and its value MUST be true. Otherwise, the property MAY be included and its default value is false
}

// PathParams add multiple path params to the provided route.
// the value may be a map[string]any with any primitive type or a slice of a single type.
// or a struct where the fields represent the values of the param.
func (r *Route) PathParams(value any) *Route {
	return r.addParams("path", value)
}

// CookieParams add multiple cookie params to the provided route.
// the value may be a map[string]any with any primitive type or a slice of a single type.
// or a struct where the fields represent the values of the param.
func (r *Route) CookieParams(value any) *Route {
	return r.addParams("cookie", value)
}

// QueryParams add multiple query params to the provided route.
// the value may be a map[string]any with any primitive type or a slice of a single type.
// or a struct where the fields represent the values of the param.
func (r *Route) QueryParams(value any) *Route {
	return r.addParams("query", value)
}

// HeaderParams add multiple header params to the provided route.
// the value may be a map[string]any with any primitive type or a slice of a single type.
// or a struct where the fields represent the values of the param.
func (r *Route) HeaderParams(value any) *Route {
	return r.addParams("header", value)
}

// addParams add a given paramType (path, query, header, cookie) to the provided route.
// the value may be a map[string]any with any primitive type or a slice of a single type.
// or a struct where the fields represent the values of the param.
func (r *Route) addParams(pType string, value any) *Route {
	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.Struct:
		typ := val.Type()
		// iterate through each field and add an example for each field. single depth only
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			fVal := val.Field(i)

			name := strings.Replace(field.Tag.Get("json"), ",omitempty", "", 1)
			desc := field.Tag.Get("desc")

			// skip unexported and ignored fields
			if name == "-" || !fVal.CanInterface() {
				continue
			}
			if name == "" {
				name = field.Name
			}
			r.AddParam(pType, name, fVal.Interface(), desc)
		}
	case reflect.Map:
		// iterate through the map and add each key/value pair. Slices are okay for adding multiple examples at the same time.
		iter := val.MapRange()
		for iter.Next() {
			k, v := iter.Key(), iter.Value()
			r.AddParam(pType, k.String(), v.Interface(), "")
		}
	default: //primitives and slices.
		// not supported
	}
	return r
}

// PathParam adds an example Path Parameter to the Route (paths)
func (r *Route) PathParam(name string, value any, desc string) *Route {
	return r.AddParam("path", name, value, desc)
}

// CookieParam adds an example Path Parameter to the Route (paths)
func (r *Route) CookieParam(name string, value any, desc string) *Route {
	return r.AddParam("cookie", name, value, desc)
}

// QueryParam adds an example Path Parameter to the Route (paths)
func (r *Route) QueryParam(name string, value any, desc string) *Route {
	return r.AddParam("query", name, value, desc)
}

// HeaderParam adds an example Path Parameter to the Route (paths)
func (r *Route) HeaderParam(name string, value any, desc string) *Route {
	return r.AddParam("header", name, value, desc)
}

// AddParam adds the given type params to the route
// pType = path, cookie, query, header
// It does not validate that the name is part of the path
// or prevent duplicate paths from being added.
// every element in value if it's a slice is added as an example.
func (r *Route) AddParam(pType, name string, value any, desc string) *Route {
	key := pType + "|" + name
	var p Param
	if r.Params == nil {
		r.Params = make(Params)
		for _, k := range parsePath(r.path) {
			r.Params["path|"+k] = Param{
				Name:     k,
				In:       "path",
				Desc:     desc,
				Examples: make(map[string]Example),
			}
		}
	}
	p, found := r.Params[key]
	if !found {
		p = Param{
			In: pType, Name: name,
			Desc:     desc,
			Examples: make(map[string]Example),
		}
		if pType == "path" {
			p.Desc = "err: not found in path"
		}

	}

typeswitch:
	switch reflect.ValueOf(value).Kind() {
	case reflect.Slice, reflect.Array:
		sliceVal := reflect.ValueOf(value)
		// check if the slices' elemental type is a primitive
		elemVal := reflect.New(sliceVal.Type().Elem()).Elem().Interface()
		if !isPrimitive(elemVal) {
			p.Desc = "err: invalid param, slice elem must be primitive"
			break
		}

		for i := 0; i < sliceVal.Len(); i++ {
			value = sliceVal.Index(i).Interface()
			exName := fmt.Sprintf("%v", value)
			if ex, ok := value.(Example); ok {
				if ex.Summary != "" {
					exName = ex.Summary
				}
				value = ex.Value
				elemVal = value
			}
			p.Examples[exName] = Example{Value: value}
		}

		if p.Schema == nil {
			s := buildSchema(elemVal)
			p.Schema = &s
		}
	case reflect.Struct:
		if ex, ok := value.(Example); ok {
			exName := ex.Summary
			ex.Summary = ""
			if exName == "" {
				exName = fmt.Sprintf("%v", value)
			}
			p.Examples[exName] = ex
			break typeswitch
		}
		fallthrough
	case reflect.Map:
		p.Desc = "err: invalid type map|struct"
	case reflect.Pointer:
		rVal := reflect.ValueOf(value).Elem()
		if rVal.Kind() == reflect.Map || rVal.Kind() == reflect.Struct {
			p.Desc = "err: invalid type map|struct"
			break
		}
		value = rVal.Interface()
		goto typeswitch
	default:
		exName := fmt.Sprintf("%v", value)
		if p.Schema == nil {
			s := buildSchema(value)
			p.Schema = &s
		}
		if !reflect.ValueOf(value).IsZero() {
			p.Examples[exName] = Example{Value: value}
		}

	}

	r.Params[key] = p
	return r
}

func isPrimitive(v any) bool {
	kind := reflect.ValueOf(v).Kind()
	if kind == reflect.Pointer {
		kind = reflect.ValueOf(v).Type().Elem().Kind()
	}
	switch kind {
	case reflect.Struct:
		_, ok := v.(Example)
		return ok
	case reflect.Slice, reflect.Array, reflect.Map, reflect.Interface, reflect.Invalid:
		return false
	default:
		return true
	}
}

// List converts the Params map to a sorted slice
func (p Params) List() []Param {
	l := make([]Param, len(p))
	i := 0
	for _, v := range p {
		l[i] = v
		i++
	}
	sort.Slice(l, func(i, j int) bool {
		if l[i].In == l[j].In {
			return l[i].Name < l[j].Name
		}
		return l[i].In < l[j].In
	})
	return l
}

func (p Params) MarshalJSON() ([]byte, error) {
	l := p.List()
	return json.Marshal(l)
}

func (p *Params) UnmarshalJSON(b []byte) error {
	l := make([]Param, 0)
	err := json.Unmarshal(b, &l)
	if err != nil {
		return err
	}
	return nil
}
