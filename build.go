package openapi

// This file should be generic settings for the openapi build options
// this needs to be put into open source so anyone can use these sdk tools to generate the openapi document

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc64"
	"log"
	"reflect"
	"strings"
	"time"
)

func NewFromJson(spec string) (api *OpenAPI, err error) {
	api = &OpenAPI{
		Paths: make(Router),
	}
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
		Tags:  make([]Tag, 0),
		Paths: make(Router),
		//ExternalDocs: &ExternalDocs{},
	}
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

const (
	Integer Type = "integer"
	Number  Type = "number"
	String  Type = "string"
	Boolean Type = "boolean"
	Object  Type = "object"
	Array   Type = "array"
)

/*
type Format string
const (
	Int32 Format = "int32"
	Int64 Format = "int64"
	Date   Format = "date"  // full-date - https://www.rfc-editor.org/rfc/rfc3339#section-5.6
	DateTime Format = "datetime" // date-time - https://www.rfc-editor.org/rfc/rfc3339#section-5.6
	Password Format = "password"
) */

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

func (o *OpenAPI) AddTags(t ...Tag) {
	o.Tags = append(o.Tags, t...)
}

// BuildSchema will create a schema object based on a given example object interface
// struct tag can be used for additional info
func buildSchema(body any) (s Schema) {
	if body == nil {
		return s
	}
	if jBody, ok := body.(JSONString); ok {
		body = jBody.ToMap()
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
		sKeys := make([]string, 0, len(keys))
		for _, k := range keys {
			sKeys = append(sKeys, k.String())
			s.Properties[k.String()] = buildSchema(value.MapIndex(k).Interface())
		}
		s.Title = GetSchemaName(sKeys)

	case reflect.Struct:
		s.Title = typ.String()
		// these are special cases for time strings
		// that may have formatting (time.Time default is RFC3339)
		switch value.Interface().(type) {
		case time.Time:
			s.Type = String
			return s
		case Time:
			s.Type = String
			return s
		case Example, *Example:
			var ex Example
			if _, ok := body.(*Example); ok {
				ex = *body.(*Example)
			} else {
				ex = body.(Example)
			}

			sch := buildSchema(ex.Value)
			s.Title = ""
			s.Type = sch.Type
			s.Desc = ex.Desc
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

			val := value.Field(i) //  the reflect.value of the struct field
			varName := field.Name // the name of the struct field
			if jsonTag != "" {
				varName = jsonTag
			}

			prop := buildSchema(val.Interface())
			prop.Desc = desc
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

// hash16 creates 16 character checksum on the string provided.
func hash16(s string) string {
	hasher := crc64.New(crc64.MakeTable(crc64.ISO))
	hasher.Write([]byte(s))
	return fmt.Sprintf("%x", hasher.Sum64())
}

// Compile the OpenAPI object by going through all
// objects and consolidating schemas and return a
// error of issues found
func (o *OpenAPI) Compile() error {
	if o.Components.Schemas == nil {
		o.Components.Schemas = make(map[string]Schema)
	}
	var errs error
	for _, r := range o.Paths {
		if r.Requests != nil {
			for k, c := range r.Requests.Content {
				if k == "invalid/json" {
					errs = errors.Join(errs, fmt.Errorf("invalid json %v request at %v: %q", r.method, r.path, c.Examples["invalid"].Value))
					continue
				}
				if c.Schema.Type != Object {
					continue
				}
				if _, found := o.Components.Schemas[c.Schema.Title]; !found {
					o.Components.Schemas[c.Schema.Title] = c.Schema
				}
				c.Schema = Schema{Ref: "#/components/schemas/" + c.Schema.Title}
				r.Requests.Content[k] = c
			}
		}
		for _, resp := range r.Responses {
			for k, c := range resp.Content {
				if k == "invalid/json" {
					errs = errors.Join(errs, fmt.Errorf("invalid json %v response at %v: %q", r.method, r.path, c.Examples["invalid"].Value))
					continue
				}
				if c.Schema.Type != Object {
					continue
				}
				if _, found := o.Components.Schemas[c.Schema.Title]; !found {
					o.Components.Schemas[c.Schema.Title] = c.Schema
				}
				c.Schema = Schema{Ref: "#/components/schemas/" + c.Schema.Title}
				resp.Content[k] = c
			}
		}

		for _, p := range r.Params {
			if strings.Contains(p.Desc, "err:") {
				errs = errors.Join(errs, fmt.Errorf("%v param %v| %v", p.In, p.Name, p.Desc))
			}

		}
	}
	return errs
}

// JSON returns the json string value for the OpenAPI object
func (o *OpenAPI) JSON() string {
	return string(o.JSONBytes())
}

func (o *OpenAPI) JSONBytes() []byte {
	b, err := json.MarshalIndent(o, "", "    ")
	if err != nil {
		log.Println(err)
	}
	return b
}
