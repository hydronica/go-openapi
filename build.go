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

type Type string
type Format int

const (
	Integer Type = "integer"
	Number  Type = "number"
	String  Type = "string"
	Boolean Type = "boolean"
	Object  Type = "object"
	Array   Type = "array"
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

// AddRoute will add a new Route to the paths object for the openapi spec
// A unique Route is need to add params, responses, and request objects
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

// AddParam adds a param object to the given unique Route
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
	if rp.Type == "" {
		param.Schema = &Schema{
			Type: String,
		}
	} else {
		param.Schema = &Schema{
			Type: rp.Type,
			//Format: rp.Format.String(),
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

// AddRequest will add a request object for the unique Route in the openapi receiver
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

	if kind == reflect.Pointer {
		if value.IsNil() { // create a new object if pointer is nil
			value = reflect.New(typ.Elem())
		}
		value = value.Elem()
		typ = value.Type()
		kind = value.Kind()
	}

	switch kind {
	case reflect.Map:
		s.Type = Object
		keys := value.MapKeys()
		if len(keys) == 0 {
			return s
		}
		if s.Properties == nil {
			s.Properties = make(Properties)
		}
		for _, k := range keys {
			s.Properties[k.String()] = buildSchema(value.MapIndex(k).Interface())
		}

	case reflect.Struct:
		// todo: when to ref rather than embed?
		// these are special cases for time strings
		// that may have formatting (time.Time default is RFC3339)
		switch value.Interface().(type) {
		case time.Time:
			s.Type = String
			return s
		case Time:
			s.Type = String
			return s
		}

		s.Type = Object
		numFields := typ.NumField()
		if s.Properties == nil {
			s.Properties = make(Properties)
		}
		for i := 0; i < numFields; i++ {
			field := typ.Field(i)
			// these are struct tags that are used in the openapi spec

			jsonTag := strings.Replace(field.Tag.Get("json"), ",omitempty", "", 1)
			desc := field.Tag.Get("desc")
			//format := field.Tag.Get("format") // used for time string formats

			// skip any fields that are not exported
			if !value.Field(i).CanInterface() || jsonTag == "-" {
				continue
			}
			// val is the reflect.value of the struct field
			val := value.Field(i)
			// the name of the struct field
			varName := field.Name
			if jsonTag != "" {
				varName = jsonTag
			}

			fieldType := typ.Field(i).Type.Kind()

			// todo do we need this pointer logic?
			if fieldType == reflect.Pointer {
				// get the value of the pointer
				va := reflect.ValueOf(val.Interface()).Elem()
				fieldType = va.Kind()
			}

			prop := buildSchema(val.Interface())
			prop.Desc = desc
			prop.Title = kind.String()
			s.Properties[varName] = prop

		}
	case reflect.Int32, reflect.Uint32:
		return Schema{Type: Integer}
	case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint64:
		return Schema{Type: Integer}
	case reflect.Float32, reflect.Float64:
		return Schema{Type: Number}
	case reflect.Bool:
		return Schema{Type: Boolean}
	case reflect.String:
		return Schema{Type: String}
	case reflect.Slice, reflect.Array:
		if k := typ.Elem().Kind(); k == reflect.Interface {
			// todo: We have a anyOf array
		} else if k == reflect.Map || k == reflect.Struct ||
			k == reflect.Array || k == reflect.Slice {
			// check the type of the first element of the array if it exists
			if value.Len() > 0 && value.IsValid() {
				prop := buildSchema(value.Index(0).Interface())
				return Schema{
					Type:  Array,
					Items: &prop,
				}
			}
		}

		// since the slice may be empty, create the child object to determine its type.
		child := reflect.New(typ.Elem()).Elem().Interface()
		prop := buildSchema(child)
		return Schema{
			Type:  Array,
			Items: &prop,
		}
	default:
		return Schema{Type: (Type)("invalid " + kind.String())}
	}

	return s
}

func reflect2Type(s string) Type {
	switch s {
	case "int":
		return Integer
	default:
		return Type(s)
	}
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
