package openapi

// This file should be generic settings for the openapi build options
// this needs to be put into open source so anyone can use these sdk tools to generate the openapi document

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	jsoniter "github.com/json-iterator/go"
)

const Default = "default"

func NewFromJson(spec string) (api *OpenAPI, err error) {
	json := jsoniter.ConfigFastest
	err = json.UnmarshalFromString(spec, &api)
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

// key is the reference name for the open api spec
type Requests map[string]RequestBody
type Params map[string]Param

type RouteParam struct {
	Name      string // unique name reference
	Desc      string // A brief description of the parameter. This could contain examples of use. CommonMark syntax MAY be used for rich text representation.
	Required  bool   // is this paramater required
	Location  string // REQUIRED. The location of the parameter. Possible values are "query", "header", "path" or "cookie".
	Type      Type   // the type for the param i.e., string, integer, float, array
	ArrayType Type   // if the Type is an array then this decribes the type for each value in the array, i.e., string, integer, Float
	Format    Format
	Example   map[string]any
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

type MIMEType string
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

type Route struct {
	Tag       string
	Desc      string
	Content   MIMEType
	ReqType   Type                  // the request type for the path i.e., array, object, string, integer
	RespType  Type                  // the response type for the path i.e., array, object, string, integer
	Responses map[string]RouteResp  // key references for responses
	Params    map[string]RouteParam // key reference for params
	Requests  map[string]RouteReq   // key reference for requests
}

type RouteResp struct {
	Code    string // response code (as a string) "200","400","302"
	Content MIMEType
	Ref     Reference // the reference name for the response object
	Array   bool      // is the response object an array
}

type RouteReq struct {
	Content MIMEType
	Ref     Reference
	Array   bool
}

type UniqueRoute struct {
	Path   string
	Method Method
}

type Tags []Tag

func (o *OpenAPI) AddTags(t Tags) {
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

	// save the given route in the spec object for later reference
	if o.Routes == nil {
		o.Routes = make(map[UniqueRoute]Route)
	}

	o.Routes[ur] = Route{
		Tag:  tag,
		Desc: desc,
	}

	p[ur.Method] = m
	o.Paths[ur.Path] = p

	return ur, nil
}

type BodyObject struct {
	MIMEType   MIMEType // the mimetype for the object
	HttpStatus string   // Any HTTP status code, '200', '201', '400' the value of 'default' can be used to cover all responses not defined
	Array      bool     // is the reference to an array
	Example    any      // the response object example used to determine the type and name of each field returned
	Desc       string   // description of the body
	Title      string   // object title
}

// NewBody is the data for mapping a request / response object to a specific route
func NewBody(mtype MIMEType, status, desc string, array bool, body any) BodyObject {
	return BodyObject{
		MIMEType:   mtype,
		HttpStatus: status,
		Array:      array,
		Example:    body,
		Desc:       desc,
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

	m.Params = append(m.Params, param)

	p[ur.Method] = m
	o.Paths[ur.Path] = p

	return nil
}

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

// AddRequest will add
func (o *OpenAPI) AddRequest(ur UniqueRoute, bo BodyObject) error {

	return nil
}

// AddResp adds response information to the api responses map
// this is used to add a response body
func (o *OpenAPI) AddResp(ur UniqueRoute, bo BodyObject) error {

	p, m, err := o.PathMethod(ur.Path, ur.Method)
	if err != nil {
		return err
	}

	var rSchema Schema
	if bo.Example != nil {
		rSchema, err = BuildSchema(bo.Title, bo.Desc, true, bo.Example)
		if err != nil {
			log.Println("error building schema for endpoint", ur.Method, ur.Path)
		}
	}

	m.Responses = Responses{
		bo.HttpStatus: Response{
			Desc: bo.Desc,
			Content: map[string]Media{
				string(bo.MIMEType): {
					Schema: rSchema,
				},
			},
		},
	}

	p[ur.Method] = m
	o.Paths[ur.Path] = p

	return nil
}

// BuildSchema will create a schema object based on a given example body interface
// if the example bool is true the body will be marshaled and added to the schema as an example
func BuildSchema(title, desc string, example bool, body any) (s Schema, err error) {

	if body == nil {
		return
	}

	if example {
		s.Example = body
	}

	typ := reflect.TypeOf(body)
	value := reflect.ValueOf(body)
	kind := typ.Kind()

	// skip any objects that are not exported, or cannot be interfaced
	if !value.CanInterface() {
		return s, nil
	}

	if kind == reflect.Pointer {
		value = value.Elem()
		typ = value.Type()
		kind = value.Kind()
	}

	s.Title = title
	s.Desc = desc

	switch kind {
	case reflect.String:
		s.Type = String.String()
	case reflect.Struct:
		//v = reflect.ValueOf(&bo.Body).Elem()
		numFields := typ.NumField()
		for i := 0; i < numFields; i++ {
			format := ""
			field := typ.Field(i)
			val := value.Field(i).Interface()
			varName := field.Name
			jsonTag := field.Tag.Get("json")
			fieldType := typ.Field(i).Type.Kind()

			if fieldType == reflect.Pointer {
				va := reflect.ValueOf(val).Elem()
				fieldType = va.Kind()
			}

			if s.Properties == nil {
				s.Properties = make(Properties)
			}
			prop := s.Properties[varName]

			simple, n, f := isSimpleType(fieldType)
			if simple {
				prop.Type = n
				format = f
			}

			prop.Desc = field.Tag.Get("description")
			if jsonTag != "" {
				varName = jsonTag
			}

			if fieldType == reflect.Struct {
				prop, err = BuildSchema("", "", false, val)
				if err != nil {
					return *s.Items, fmt.Errorf("error building struct field schema %w", err)
				}
				prop.Type = Object.String()

			}

			if fieldType == reflect.Slice {
				prop.Type = Array.String()
				t := reflect.TypeOf(val).Elem()
				fieldKind := t.Kind()
				obj := reflect.New(t).Interface()
				if fieldKind == reflect.Struct {
					items, err := BuildSchema("", "", false, obj)
					if err != nil {
						return *s.Items, fmt.Errorf("error building array field schema %w", err)
					}
					prop.Items = &items
					prop.Items.Type = Object.String()
				}
				simple, name, format := isSimpleType(fieldKind)
				if simple {
					prop.Items = &Schema{
						Type:   name,
						Format: format,
					}
				}

			}

			prop.Format = format

			s.Properties[varName] = prop

		}
	case reflect.Array:
		s.Type = Array.String()
	case reflect.Slice:
		s.Type = Array.String()
		elem := reflect.TypeOf(body).Elem()
		slicek := elem.Kind()
		obj := reflect.New(elem).Interface()
		if slicek == reflect.Struct {
			prop, err := BuildSchema("", "", false, obj)
			if err != nil {
				return *s.Items, fmt.Errorf("error building schema %w", err)
			}
			prop.Type = Object.String()
			s.Items = &prop
		}
		simple, name, format := isSimpleType(slicek)
		if simple {
			s.Items = &Schema{
				Type:   name,
				Format: format,
			}
		}

	}

	return s, nil
}

func isSimpleType(t reflect.Kind) (simple bool, kind, format string) {
	switch t {
	case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint64:
		return true, Integer.String(), Int64.String()
	case reflect.Int32, reflect.Uint32:
		return true, Integer.String(), Int32.String()
	case reflect.Float32, reflect.Float64:
		return true, Float.String(), Float.String()
	case reflect.String:
		return true, String.String(), ""
	default:
		return false, t.String(), ""
	}
}

// JSON returns the json string value for the OpenAPI object
func (o *OpenAPI) JSON() []byte {
	json := jsoniter.ConfigFastest
	b, _ := json.Marshal(o)
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
