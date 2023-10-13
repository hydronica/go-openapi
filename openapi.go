package openapi

import (
	"strconv"
)

// OpenAPI represents the definition of the openapi specification 3.0.3
type OpenAPI struct {
	Version string   `json:"openapi"`           // the  semantic version number of the OpenAPI Specification version
	Tags    []Tag    `json:"tags,omitempty"`    // A list of tags used by the specification with additional metadata
	Servers []Server `json:"servers,omitempty"` // Array of Server Objects, which provide connectivity information to a target server.
	Paths   Paths    `json:"paths"`             // REQUIRED. Map of uri paths mapped to methods i.e., get, put, post, delete
	Info    Info     `json:"info"`              // REQUIRED. Provides metadata about the API. The metadata MAY be used by tooling as required.
	//Components   Components    `json:"components,omitempty"`   // reuseable components not used here
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"` //Additional external documentation.
}

// Operation describes a single API operation on a path.
type Operation struct {
	Tags         []string      `json:"tags,omitempty"`
	Summary      string        `json:"summary,omitempty"`
	Desc         string        `json:"description,omitempty"`
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"`
	OperationID  string        `json:"operationId,omitempty"`
	Params       []Param       `json:"parameters,omitempty"`  // A list of parameters that are applicable for this operation.
	RequestBody  *RequestBody  `json:"requestBody,omitempty"` // The request body applicable for this operation.
	Responses    Responses     `json:"responses,omitempty"`   // key 200,400 REQUIRED. The list of possible responses as they are returned from executing this operation.
}

// Paths holds the relative paths to the individual endpoints and their operations.
type Paths map[string]OperationMap //[url_path]OperationMap

// OperationMap describes the operations available on a single path.
type OperationMap map[Method]Operation // map of methods to a openAPI Operation object

// Code a valid https status such as '200', '201', '400', 'default'
type Code int

const DefaultStatus Code = 0

func (c Code) MarshalText() ([]byte, error) {
	if c == DefaultStatus {
		return []byte("default"), nil
	}
	return []byte(strconv.Itoa(int(c))), nil
}

func (c *Code) UnmarshalText(b []byte) error {
	if string(b) == "default" {
		*c = DefaultStatus
		return nil
	}
	i, err := strconv.Atoi(string(b))
	*c = Code(i)
	return err
}

type MIMEType string
type Content map[MIMEType]Media

// Responses for the expected responses of an operation, maps a HTTP response code to the expected response.
type Responses map[Code]Response

// Response describes a single response from an API Operation
type Response struct {
	Status   Code     `json:"-"`
	MimeType MIMEType `json:"-"`

	Desc    string  `json:"description,omitempty"` // A short description of the response. CommonMark syntax MAY be used for rich text representation.
	Content Content `json:"content,omitempty"`     // A map containing descriptions of potential response payloads. The key is a media type or media type range and the value describes it.
}

type Media struct {
	Schema   Schema              `json:"schema,omitempty"`   // The schema defining the content of the request, response, or parameter.
	Example  any                 `json:"example,omitempty"`  // Example of the media type. The example object SHOULD be in the correct format as specified by the media type. The example field is mutually exclusive of the examples field. Furthermore, if referencing a schema which contains an example, the example value SHALL override the example provided by the schema.
	Examples map[string]Example  `json:"examples,omitempty"` // Examples of the media type. Each example object SHOULD match the media type and specified schema if present. The examples field is mutually exclusive of the example field. Furthermore, if referencing a schema which contains an example, the examples value SHALL override the example provided by the schema.
	Encoding map[string]Encoding `json:"encoding,omitempty"` // A map between a property name and its encoding information. The key, being the property name, MUST exist in the schema as a property.
}

type Encoding struct {
	ContentType string `json:"contentType,omitempty"` // The Content-Type for encoding a specific property.
	// headers  map[string]headerObject :  not implemented needed if media is multipart
	Style string `json:"style"` // Describes how a specific property value will be serialized depending on its type.
	// explode       bool not implemented needed if media is application/x-www-form-urlencoded
	// allowReserved bool not implemented needed if media is application/x-www-form-urlencoded
}

// Example object MAY be extended with Specification Extensions.
type Example struct {
	Summary       string `json:"summary,omitempty"`       // Short description for the example.
	Desc          string `json:"description,omitempty"`   // Long description for the example. CommonMark syntax MAY be used for rich text representation.
	ExternalValue string `json:"externalValue,omitempty"` // A URL that points to the literal example. This provides the capability to reference examples that cannot easily be included in JSON or YAML documents. The value field and externalValue field are mutually exclusive.
	Value         any    `json:"value"`                   // Embedded literal example. The value field and externalValue field are mutually exclusive. To represent examples of media types that cannot naturally represented in JSON or YAML, use a string value to contain the example, escaping where necessary.
}

// RequestBody describes a single request body.
type RequestBody struct {
	Desc     string  `json:"description,omitempty"` // A brief description of the request body. This could contain examples of use. CommonMark syntax MAY be used for rich text representation.
	Content  Content `json:"content,omitempty"`     // REQUIRED. The content of the request body. The key is a media type or media type range and the value describes it. For requests that match multiple keys, only the most specific key is applicable. e.g. text/plain overrides text/*
	Required bool    `json:"required,omitempty"`    // Determines if the request body is required in the request. Defaults to false.
}

// Schema Object allows the definition of input and output data types
// These types can be objects, but also primitives and arrays
// This object is an extended subset of the JSON Schema Specification
type Schema struct {
	AddProperties *Schema       `json:"additionalProperties,omitempty"` // To define a dictionary, use type: object and use the additionalProperties keyword to specify the type of values in key/value pairs.
	Title         string        `json:"title,omitempty"`                // Title?
	Desc          string        `json:"description,omitempty"`          // CommonMark syntax MAY be used for rich text representation.
	Type          string        `json:"type,omitempty"`                 // Value MUST be a string. Multiple types via an array are not supported.
	Format        string        `json:"format,omitempty"`               // See Data Type Formats for further details. While relying on JSON Schema's defined formats, the OAS offers a few additional predefined formats.
	Items         *Schema       `json:"items,omitempty"`                // Value MUST be an object and not an array. Inline or referenced schema MUST be of a Schema Object and not a standard JSON Schema. items MUST be present if the type is array.
	Properties    Properties    `json:"properties,omitempty"`           // Property definitions MUST be a Schema Object and not a standard JSON Schema (inline or referenced).
	Example       any           `json:"example,omitempty"`              // A free-form property to include an example of an instance for this schema. To represent examples that cannot be naturally represented in JSON or YAML, a string value can be used to contain the example with escaping where necessary.
	ExternalDocs  *ExternalDocs `json:"externalDocs,omitempty"`         // Additional external documentation for this schema.
}

type Properties map[string]Schema

// Param see https://swagger.io/docs/specification/describing-parameters/
// - Path /user/{id}
// - Query /user?role=admin
// - header X-MyHeader: Value
// - cookie
type Param struct {
	Name     string             `json:"name,omitempty"`        // REQUIRED. The name of the parameter. Parameter names are case sensitive.
	Desc     string             `json:"description,omitempty"` // A brief description of the parameter. This could contain examples of use. CommonMark syntax MAY be used for rich text representation.
	Style    string             `json:"style,omitempty"`       // Describes how the parameter value will be serialized depending on the type of the parameter value. Default values (based on value of in): for query - form; for path - simple; for header - simple; for cookie - form.
	In       string             `json:"in"`                    // REQUIRED. The location of the parameter. Possible values are "query", "header", "path" or "cookie".
	Schema   *Schema            `json:"schema,omitempty"`      // The schema defining the type used for the parameter.
	Examples map[string]Example `json:"examples"`              // Examples of the parameter’s potential value. Each example SHOULD contain a value in the correct format as specified in the parameter encoding.
	Required bool               `json:"required"`              // Determines whether this parameter is mandatory. If the parameter location is "path", this property is REQUIRED and its value MUST be true. Otherwise, the property MAY be included and its default value is false
}

type Info struct {
	Title   string   `json:"title"`                    // REQUIRED. The title of the API.
	Version string   `json:"version" required:"true"`  // REQUIRED. The version of the OpenAPI document (which is distinct from the OpenAPI Specification version or the API implementation version).
	Desc    string   `json:"description"`              // A short description of the API. CommonMark syntax MAY be used for rich text representation.
	Terms   string   `json:"termsOfService,omitempty"` // A URL to the Terms of Service for the API. MUST be in the format of a URL.
	Contact *Contact `json:"contact,omitempty"`        // The contact information for the exposed API.
	License *License `json:"license,omitempty"`        // The license information for the exposed API.
}

type Contact struct {
	Name  string `json:"name"`  // The identifying name of the contact person/organization.
	URL   string `json:"url"`   // The URL pointing to the contact information. MUST be in the format of a URL.
	Email string `json:"email"` // The email address of the contact person/organization. MUST be in the format of an email address.
}

type License struct {
	Name string `json:"name"` // REQUIRED. The license name used for the API.
	URL  string `json:"url"`  // A URL to the license used for the API. MUST be in the format of a URL.
}

type Server struct {
	URL  string               `json:"url"`                 // REQUIRED. A URL to the target host. This URL supports Server Variables and MAY be relative, to indicate that the host location is relative to the location where the OpenAPI document is being served. Variable substitutions will be made when a variable is named in {brackets}.
	Desc string               `json:"description"`         // An optional string describing the host designated by the URL. CommonMark syntax MAY be used for rich text representation.
	Vars map[string]ServerVar `json:"variables,omitempty"` // A map between a variable name and its value. The value is used for substitution in the server's URL template.
}

type ServerVar struct {
	Enum    []string `json:"enum"`        // An enumeration of string values to be used if the substitution options are from a limited set. The array SHOULD NOT be empty.
	Default string   `json:"default"`     // REQUIRED. The default value to use for substitution, which SHALL be sent if an alternate value is not supplied. Note this behavior is different than the Schema Object's treatment of default values, because in those cases parameter values are optional. If the enum is defined, the value SHOULD exist in the enum's values.
	Desc    string   `json:"description"` // An optional description for the server variable. CommonMark syntax MAY be used for rich text representation.
}

type Tag struct {
	Name         string        `json:"name" required:"true"`   // REQUIRED. The name of the tag.
	Desc         string        `json:"description"`            // A short description for the tag. CommonMark syntax MAY be used for rich text representation.
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"` // Additional external documentation for this tag.
}

type ExternalDocs struct {
	Desc string `json:"description,omitempty""`        // A short description of the target documentation. CommonMark syntax MAY be used for rich text representation.
	URL  string `json:"url,omitempty" required:"true"` // REQUIRED. The URL for the target documentation. Value MUST be in the format of a URL.
}
