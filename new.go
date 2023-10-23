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
	r, found := o.Paths[key]
	if !found {
		r = &Route{path: path, method: method}
		o.Paths[key] = r
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

	return r.AddExample(m)
}

// AddExample takes a struct and adds a json Content to the Response
func (r Response) AddExample(i any) Response {
	m := r.Content[Json]
	m.AddExample(i)
	return r
}

// AddExample will add an example object by
// creating a schema based on the object i passed in.
// The Example name will be the title of the Schema
// and any description from added to the example as well.
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

	return r.AddExample(m)
}

func (r RequestBody) AddExample(i any) RequestBody {
	m := r.Content[Json]
	m.AddExample(i)
	return r
}

func (r *Route) AddRequest(req RequestBody) *Route {
	r.Requests = &req
	return r
}

// AddParam adds the given type params to the route
// pType = path, cookie, query, header
// It does not validate that the name is part of the path
// or prevent duplicate paths from being added.
// every element in value if it's a slice is added as an example.
func (r *Route) AddParam(pType string, name string, value any) *Route {
	key := pType + "|" + name
	var p Param
	if r.Params == nil {
		r.Params = make(Params)
		for _, k := range parsePath(r.path) {
			r.Params["path|"+k] = Param{
				Name:     k,
				In:       "path",
				Examples: make(map[string]Example),
			}
		}
	}

	p, found := r.Params[key]
	if !found {
		p = Param{
			In: pType, Name: name,
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
		if p.Schema == nil {
			s := buildSchema(elemVal)
			p.Schema = &s
		}
		for i := 0; i < sliceVal.Len(); i++ {
			value = sliceVal.Index(i).Interface()
			exName := fmt.Sprintf("%v", value)
			p.Examples[exName] = Example{Value: value}
		}
	case reflect.Map, reflect.Struct:
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
		p.Examples[exName] = Example{Value: value}
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
	case reflect.Struct, reflect.Slice, reflect.Array, reflect.Map, reflect.Interface:
		return false
	default:
		return true
	}
}

// AddParams add a given paramType (path, query, header, cookie) to the provided route.
// the value may be a map[string]any with any primitive type or a slice of a single type.
// or a struct where the fields represent the values of the param.
func (r *Route) AddParams(pType string, value any) *Route {
	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.Struct:
		typ := val.Type()
		// iterate through each field and add an example for each field. single depth only
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			fVal := val.Field(i)

			name := strings.Replace(field.Tag.Get("json"), ",omitempty", "", 1)
			//desc := field.Tag.Get("desc")

			// skip unexported and ignored fields
			if name == "-" || !fVal.CanInterface() {
				continue
			}
			if name == "" {
				name = field.Name
			}
			r.AddParam(pType, name, fVal.Interface())
		}
	case reflect.Map:
		// iterate through the map and add each key/value pair. Slices are okay for adding multiple examples at the same time.
		iter := val.MapRange()
		for iter.Next() {
			k, v := iter.Key(), iter.Value()
			r.AddParam(pType, k.String(), v.Interface())
		}
	default: //primitives and slices.
		// not supported
	}
	return r
}
