package openapi

// This file should be generic settings for the openapi build options
// this needs to be put into open source so anyone can use these sdk tools to generate the openapi document

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

const Default = "default"

func NewFromJson(spec string) (api *OpenAPI, err error) {
	err = json.Unmarshal([]byte(spec), &api)
	if err != nil {
		return nil, fmt.Errorf("error with unmarshal %w", err)
	}
	return api, nil
}

func New(title, version, description string) *OpenAPI {
	return &OpenAPI{
		Version: "3.0.3",
		Info: Info{
			Title:   title,
			Version: version,
			Desc:    description,
		},
		Tags:         make([]Tag, 0),
		Paths:        map[string]OperationMap{}, // a map of methods mapped to operations i.e., get, put, post, delete
		ExternalDocs: &ExternalDocs{},
	}
}

type Requests map[string]RequestBody // key is the reference name for the open api spec
type Params map[string]Param

type RouteParam struct {
	Name      string          // unique name reference
	Desc      string          // A brief description of the parameter. This could contain examples of use. CommonMark syntax MAY be used for rich text representation.
	Required  bool            // is this paramater required
	Location  string          // REQUIRED. The location of the parameter. Possible values are "query", "header", "path" or "cookie".
	Type      Type            // the type for the param i.e., string, integer, float, array
	ArrayType Type            // if the Type is an array then this decribes the type for each value in the array, i.e., string, integer, Float
	Format    Format          // object format
	Examples  []ExampleObject // example param values
}

type Method string

const (
	GET     Method = "get"
	PUT     Method = "put"
	POST    Method = "post"
	DELETE  Method = "delete"
	OPTIONS Method = "options"
	HEAD    Method = "head"
	PATCH   Method = "patch"
	TRACE   Method = "trace"
)

type Type int
type Format int

const (
	Integer Type = iota + 1
	Number
	String
	Boolean
	Object
	Array
)

const (
	Int32 Format = iota + 1
	Int64
	Float
	Double
	Byte     // base64 encoded characters
	Binary   // any sequence of octets
	Date     // full-date - https://www.rfc-editor.org/rfc/rfc3339#section-5.6
	DateTime // date-time - https://www.rfc-editor.org/rfc/rfc3339#section-5.6
	Password
)

func (t Type) String() string {
	switch t {
	case Integer:
		return "integer"
	case Number:
		return "number"
	case String:
		return "string"
	case Boolean:
		return "boolean"
	case Object:
		return "object"
	case Array:
		return "array"
	}
	return ""
}

func (f Format) String() string {
	switch f {
	case Int32:
		return "int32"
	case Int64:
		return "int64"
	case Float:
		return "float"
	case Double:
		return "double"
	case Byte:
		return "byte"
	case Binary:
		return "binary"
	case Date:
		return "date"
	case DateTime:
		return "dateTime"
	case Password:
		return "password"
	}
	return ""
}

type Reference string

// common media types
const (
	Json    MIMEType = "application/json"
	Xml     MIMEType = "application/xml"
	Text    MIMEType = "text/plain"
	General MIMEType = "application/octet-stream"
	Html    MIMEType = "text/html"
	XForm   MIMEType = "application/x-www-form-urlencoded"
	Jscript MIMEType = "application/javascript"
	Form    MIMEType = "multipart/form-data"
)

type UniqueRoute struct {
	Path   string
	Method Method
}

func (o *OpenAPI) AddTags(t ...Tag) {
	o.Tags = append(o.Tags, t...)
}

func (o *OpenAPI) AddTag(tag, description string) {
	o.Tags = append(o.Tags, Tag{
		Name: tag,
		Desc: description,
	})
}

// AddRoute will add a new route to the paths object for the openapi spec
// A unique route is need to add params, responses, and request objects
func (o *OpenAPI) AddRoute(path, method, tag, desc, summary string) (ur UniqueRoute, err error) {
	if tag == "" {
		tag = Default
	}
	if path == "" || method == "" {
		return ur, fmt.Errorf("path and method cannot be an empty string")
	}

	ur = UniqueRoute{
		Path:   path,
		Method: Method(method),
	}

	// initialize the paths if nil
	if o.Paths == nil {
		o.Paths = make(Paths)
	}

	p, found := o.Paths[ur.Path]
	if !found {
		o.Paths[ur.Path] = make(OperationMap)
		p = o.Paths[ur.Path]
	}

	m := p[ur.Method]
	m.Desc = desc
	m.Tags = append(m.Tags, tag)
	m.OperationID = string(ur.Method) + "_" + ur.Path

	p[ur.Method] = m
	o.Paths[ur.Path] = p

	return ur, nil
}

type ExampleObject struct {
	Example any    // Example value
	Name    string // object name
	Summary string // short description summary
	Desc    string // full (long) description
}

type BodyObject struct {
	MIMEType   MIMEType        // the mimetype for the object
	HttpStatus Code            // Any HTTP status code, '200', '201', '400' the value of 'default' can be used to cover all responses not defined
	Examples   []ExampleObject // the response object examples used to determine the type and name of each field returned
	Desc       string          // description of the body
	Title      string          // object title
}

// NewRespBody is a helper function to create a response body object
// example is a go object to represent the body
func NewRespBody(mtype MIMEType, status Code, desc string, examples []ExampleObject) BodyObject {
	return BodyObject{
		MIMEType:   mtype,
		HttpStatus: status,
		Examples:   examples,
		Desc:       desc,
	}
}

// NewReqBody is a helper function to create a request body object
// example is a go object to represent the body
func NewReqBody(mtype MIMEType, desc string, examples []ExampleObject) BodyObject {
	return BodyObject{
		MIMEType: mtype,
		Examples: examples,
		Desc:     desc,
	}
}

// AddParam adds a param object to the given unique route
func (o *OpenAPI) AddParam(ur UniqueRoute, rp RouteParam) error {
	if rp.Name == "" || rp.Location == "" {
		return fmt.Errorf("param name and location are required to add param")
	}
	p, found := o.Paths[ur.Path]
	if !found {
		return fmt.Errorf("could not find path to add param %v", ur)
	}
	m, found := p[ur.Method]
	if !found {
		return fmt.Errorf("could not find method to add param %v", ur)
	}
	param := Param{
		Name: rp.Name,
		Desc: rp.Desc,
		In:   rp.Location,
	}

	if rp.Location == "path" {
		param.Required = true
		param.Style = "simple"
	}

	// if no type is given defaults to string
	if rp.Type == 0 {
		param.Schema = &Schema{
			Type: String.String(),
		}
	} else {
		param.Schema = &Schema{
			Type:   rp.Type.String(),
			Format: rp.Format.String(),
		}
	}

	param.Examples = make(map[string]Example)

	for _, e := range rp.Examples {
		param.Examples[e.Name] = Example{
			Value:   e.Example,
			Summary: e.Summary,
			Desc:    e.Desc,
		}
	}

	m.Params = append(m.Params, param)

	p[ur.Method] = m
	o.Paths[ur.Path] = p

	return nil
}

// PathMethod is a helper method to pull the operation map out of the openapi paths map
// then it pulls the operation struct out of the operation map for a fast reference to the operation struct
func (o *OpenAPI) PathMethod(path string, method Method) (om OperationMap, op Operation, err error) {
	om, found := o.Paths[path]
	if !found {
		return om, op, fmt.Errorf("could not find path to add param %v", path)
	}
	op, found = om[method]
	if !found {
		return om, op, fmt.Errorf("could not find method to add param %v", method)
	}
	return om, op, nil
}

// AddRequest will add a request object for the unique route in the openapi receiver
// adds an example and schema to the request body
func (o *OpenAPI) AddRequest(ur UniqueRoute, bo BodyObject) error {
	om, op, err := o.PathMethod(ur.Path, ur.Method)
	if err != nil {
		return err
	}

	examples := make(map[string]Example)
	var rSchema Schema
	if len(bo.Examples) > 0 {
		rSchema = buildSchema(bo.Examples[0].Example)
		for i, e := range bo.Examples {
			if e.Name == "" {
				e.Name = fmt.Sprintf("example: %d", i)
			}
			examples[e.Name] = Example{
				Summary: e.Summary,
				Desc:    e.Desc,
				Value:   e.Example,
			}
		}
	}

	if op.RequestBody == nil {
		op.RequestBody = &RequestBody{
			Content: make(Content),
		}
	}

	r := op.RequestBody.Content[bo.MIMEType]
	r.Examples = examples
	r.Schema = rSchema
	r.Schema.Desc = bo.Desc
	op.RequestBody.Content[bo.MIMEType] = r
	om[ur.Method] = op
	o.Paths[ur.Path] = om

	return nil
}

// AddResponse adds response information to the api responses map which is part of the paths map
// adds an example and schema to the response body
func (o *OpenAPI) AddResponse(ur UniqueRoute, bo BodyObject) error {
	om, op, err := o.PathMethod(ur.Path, ur.Method)
	if err != nil {
		return err
	}
	examples := make(map[string]Example)
	var rSchema Schema
	if len(bo.Examples) > 0 {
		rSchema = buildSchema(bo.Examples[0].Example)
		for i, e := range bo.Examples {
			if e.Name == "" {
				e.Name = fmt.Sprintf("example: %d", i+1)
			}
			examples[e.Name] = Example{
				Summary: e.Summary,
				Desc:    e.Desc,
				Value:   e.Example,
			}
		}
	}

	if op.Responses == nil {
		op.Responses = make(Responses)
	}

	r := op.Responses[bo.HttpStatus]
	r.Desc = bo.Desc
	r.Content = Content{
		bo.MIMEType: {
			Schema:   rSchema,
			Examples: examples,
		},
	}

	op.Responses[bo.HttpStatus] = r
	om[ur.Method] = op
	o.Paths[ur.Path] = om

	return nil
}

// BuildSchema will create a schema object based on a given example object interface
// struct tag can be used for additional info
func buildSchema(body any) (s Schema) {
	if body == nil {
		return s
	}

	value := reflect.ValueOf(body)
	typ := reflect.TypeOf(body)
	kind := typ.Kind()

	// skip any objects that are not exported, or cannot be interfaced
	if !value.CanInterface() {
		return s
	}

	if kind == reflect.Pointer {
		value = value.Elem()
		if !value.IsValid() {
			return s
		}
		typ = value.Type()
		kind = value.Kind()
	}

	switch kind {
	case reflect.Int32, reflect.Uint32:
		s.Type = Integer.String()
		s.Format = Int32.String()
	case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint64:
		s.Type = Integer.String()
		s.Format = Int64.String()
	case reflect.Float32, reflect.Float64:
		s.Type = Number.String()
		s.Format = Float.String()
	case reflect.Bool:
		s.Type = Boolean.String()
	case reflect.String:
		s.Type = String.String()

	case reflect.Map:
		s.Type = Object.String()
		keys := value.MapKeys()
		if len(keys) == 0 {
			return s
		}
		// build out the map keys as a property schemas for openapi
		for _, k := range keys {
			v := value.MapIndex(k)
			field := k.String()
			if s.Properties == nil {
				s.Properties = make(Properties)
			}

			schema := buildSchema(v.Interface())

			s.Properties[field] = schema
		}

	case reflect.Struct:
		// these are special cases for time strings
		// that may have formatting (time.Time default is RFC3339)
		switch x := value.Interface().(type) {
		case time.Time:
			s.Type = String.String()
			s.Format = time.RFC3339
			/*if f, ok := tags["format"]; ok && f != "" {
				s.Format = tags["format"]
			}*/
			return s
		case Time:
			s.Type = String.String()
			s.Format = x.Format
			return s
		}

		s.Type = Object.String()
		numFields := typ.NumField()
		for i := 0; i < numFields; i++ {
			field := typ.Field(i)
			// these are struct tags that are used in the openapi spec
			tags := map[string]string{
				"json":        field.Tag.Get("json"),   // used for the field name
				"description": field.Tag.Get("desc"),   // used for field description
				"format":      field.Tag.Get("format"), // used for time string formats
			}

			// skip any fields that are not exported
			if !value.Field(i).CanInterface() {
				continue
			}
			// val is the reflect.value of the struct field
			val := value.Field(i)
			// the name of the struct field
			varName := field.Name
			// the json tag string value
			// in go the json tag - is a skipped field (not output to json)
			if tags["json"] == "-" {
				continue
			}
			if tags["json"] != "" {
				// ,omitempty is a go json template option to ignore the field if it has a zero value
				varName = strings.Replace(tags["json"], ",omitempty", "", 1)
			}
			fieldType := typ.Field(i).Type.Kind()

			if fieldType == reflect.Pointer {
				// get the value of the pointer
				va := reflect.ValueOf(val.Interface()).Elem()
				fieldType = va.Kind()
			}

			if s.Properties == nil {
				s.Properties = make(Properties)
			}
			prop := s.Properties[varName]
			prop.Desc = tags["description"]

			if val.IsValid() {
				i := val.Interface()
				prop = buildSchema(i)
				s.Properties[varName] = prop
			}
		}

	case reflect.Slice, reflect.Array:
		prop := Schema{}
		s.Type = Array.String()
		if value.Len() > 0 && value.IsValid() {
			obj := value.Index(0).Interface()
			prop = buildSchema(obj)

		}
		s.Items = &prop
	default:
		fmt.Println("SHOULD NEVER GET HERE")
	}

	return s
}

type TypeInfo struct {
	Simple bool   // is the type a primitive type i.e., int, float, string, bool
	Type   string // the type name if this is a primitive type
	Format string // a format for the given type such as int64 int32 float
}

// JSON returns the json string value for the OpenAPI object
func (o *OpenAPI) JSON() []byte {
	b, _ := json.Marshal(o)
	b, _ = JSONRemarshal(b)
	return b
}

// This will re-marshal the bytes so that the map key fields are sorted accordingly.
func JSONRemarshal(bytes []byte) ([]byte, error) {
	var ifce interface{}
	err := json.Unmarshal(bytes, &ifce)
	if err != nil {
		return nil, err
	}
	return json.Marshal(ifce)
}
