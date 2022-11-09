package spec

type OpenAPI struct {
	Version string   `json:"openapi"`           // the  semantic version number of the OpenAPI Specification version
	Tags    []Tag    `json:"tags,omitempty"`    // A list of tags used by the specification with additional metadata
	Servers []Server `json:"servers,omitempty"` // Array of Server Objects, which provide connectivity information to a target server.
	Paths   Paths    `json:"paths"`             // REQUIRED. Map of uri paths mapped to methods i.e., get, put, post, delete
	Info    Info     `json:"info"`              // REQUIRED. Provides metadata about the API. The metadata MAY be used by tooling as required.
	//Components   Components    `json:"components,omitempty"`   // reuseable components not used here
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"` //Additional external documentation.

	// non OpenAPI external reference for simplified routes
	Routes map[UniqueRoute]Route `json:"-"`
}

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

type Paths map[string]OperationMap
type OperationMap map[Method]Operation // map of methods to a openAPI Operation object
type Responses map[string]Response
type Response struct {
	Desc    string           `json:"description,omitempty"`
	Content map[string]Media `json:"content,omitempty"`
}

type Media struct {
	Schema Schema `json:"schema"`
}

// reusable reference objects
type Components struct {
	Schemas       Schemas                `json:"schema,omitempty"`
	RequestBodies map[string]RequestBody `json:"requestBodies,omitempty"`
	Params        map[string]Param       `json:"params,omitempty"`
}

type Schemas map[string]Schema

type RequestBody struct {
	Desc     string             `json:"description,omitempty"`
	Content  map[MIMEType]Media `json:"content,omitempty"`
	Required *bool              `json:"required,omitempty"`
}

type Schema struct {
	AddProperties *Schema    `json:"additionalProperties"`
	Title         string     `json:"title,omitempty"`
	Desc          string     `json:"description,omitempty"`
	Type          string     `json:"type,omitempty"`
	Items         *Schema    `json:"items"`
	Properties    Properties `json:"properties,omitempty"`
}

type Properties map[string]Property

type Ref struct {
	Ref string `json:"$ref"`
}

type Param struct {
	Name string `json:"name,omitempty"`        // REQUIRED. The name of the parameter. Parameter names are case sensitive.
	Desc string `json:"description,omitempty"` // A brief description of the parameter. This could contain examples of use. CommonMark syntax MAY be used for rich text representation.
	In   string `json:"in"`                    // REQUIRED. The location of the parameter. Possible values are "query", "header", "path" or "cookie".
}

type Property struct {
	Type       string `json:"type"`
	Format     string `json:"format"`
	Desc       string `json:"description"`
	Properties `json:"properties"`
	Items      *Schema `json:"items"`
}

type Info struct {
	Title   string   `json:"title"`                   // REQUIRED. The title of the API.
	Desc    string   `json:"description"`             // A short description of the API. CommonMark syntax MAY be used for rich text representation.
	Terms   string   `json:"termsOfService"`          // A URL to the Terms of Service for the API. MUST be in the format of a URL.
	Contact *Contact `json:"contact"`                 // The contact information for the exposed API.
	License *License `json:"license"`                 // The license information for the exposed API.
	Version string   `json:"version" required:"true"` // REQUIRED. The version of the OpenAPI document (which is distinct from the OpenAPI Specification version or the API implementation version).
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
	URL  string               `json:"url"`         // REQUIRED. A URL to the target host. This URL supports Server Variables and MAY be relative, to indicate that the host location is relative to the location where the OpenAPI document is being served. Variable substitutions will be made when a variable is named in {brackets}.
	Desc string               `json:"description"` // An optional string describing the host designated by the URL. CommonMark syntax MAY be used for rich text representation.
	Vars map[string]ServerVar `json:"variables"`   // A map between a variable name and its value. The value is used for substitution in the server's URL template.
}

type ServerVar struct {
	Enum    []string `json:"enum"`        // An enumeration of string values to be used if the substitution options are from a limited set. The array SHOULD NOT be empty.
	Default string   `json:"default"`     // REQUIRED. The default value to use for substitution, which SHALL be sent if an alternate value is not supplied. Note this behavior is different than the Schema Object's treatment of default values, because in those cases parameter values are optional. If the enum is defined, the value SHOULD exist in the enum's values.
	Desc    string   `json:"description"` // An optional description for the server variable. CommonMark syntax MAY be used for rich text representation.
}

type Tag struct {
	Name         string        `json:"name"`
	Desc         string        `json:"description"`
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"`
}

type ExternalDocs struct {
	Desc string `json:"description"`
	URL  string `json:"url" required:"true"`
}
