package openapi

import (
	"strconv"
)

// OpenAPI represents the definition of the openapi specification 3.0.3
type OpenAPI struct {
	Version      string        `json:"openapi"`                // the  semantic version number of the OpenAPI Specification version
	Servers      []Server      `json:"servers,omitempty"`      // Array of Server Objects, which provide connectivity information to a target server.
	Info         Info          `json:"info"`                   // REQUIRED. Provides metadata about the API. The metadata MAY be used by tooling as required.
	Tags         []Tag         `json:"tags,omitempty"`         // A list of tags used by the specification with additional metadata
	Paths        Router        `json:"paths"`                  // key= path|method
	Components   Components    `json:"components,omitempty"`   // reuseable components
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"` //Additional external documentation.
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

type Tag struct {
	Name         string        `json:"name" required:"true"`   // REQUIRED. The name of the tag.
	Desc         string        `json:"description"`            // A short description for the tag. CommonMark syntax MAY be used for rich text representation.
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"` // Additional external documentation for this tag.
}

type ExternalDocs struct {
	Desc string `json:"description,omitempty"`         // A short description of the target documentation. CommonMark syntax MAY be used for rich text representation.
	URL  string `json:"url,omitempty" required:"true"` // REQUIRED. The URL for the target documentation. Value MUST be in the format of a URL.
}

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

type Media struct {
	Schema Schema `json:"schema,omitempty"` // The schema defining the content of the request, response, or parameter.
	// Examples of the media type. Each example object SHOULD match the media type and specified schema if present. The examples field is mutually exclusive of the example field. Furthermore, if referencing a schema which contains an example, the examples value SHALL override the example provided by the schema.
	Examples map[string]Example `json:"examples,omitempty"`

	// NOT Supported:
	//Example of the media type. The example object SHOULD be in the correct format as specified by the media type. The example field is mutually exclusive of the examples field. Furthermore, if referencing a schema which contains an example, the example value SHALL override the example provided by the schema.
	//Example  any                 `json:"example,omitempty"` -> uses examples even for one example
	//A map between a property name and its encoding information. The key, being the property name, MUST exist in the schema as a property.
	//Encoding map[string]Encoding `json:"encoding,omitempty"`
}

type Components struct {
	Schemas map[string]Schema `json:"schemas,omitempty"`

	//NOT implemented
	/*
		Parameters []Params
		SecuritySchemes struct{}
		RequestBodies []RequestBody
		Responses Responses
		Headers []Params
		Examples []Example
		Links []string
		Callbacks struct{} */
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
	Summary string `json:"summary,omitempty"`     // Short description for the example.
	Desc    string `json:"description,omitempty"` // Long description for the example. CommonMark syntax MAY be used for rich text representation.
	//ExternalValue string `json:"externalValue,omitempty"` // A URL that points to the literal example. This provides the capability to reference examples that cannot easily be included in JSON or YAML documents. The value field and externalValue field are mutually exclusive.
	Value any `json:"value"` // Embedded literal example. The value field and externalValue field are mutually exclusive. To represent examples of media types that cannot naturally represented in JSON or YAML, use a string value to contain the example, escaping where necessary.
}

// Schema Object defines data types. objects (structs), maps, primitives and arrays
// This object is an extended subset of the JSON Schema Specification
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

type Properties map[string]Schema
