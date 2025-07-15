package openapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

type Time struct {
	time.Time
	Format string
}

// custom marshal for formatting time based on the Format struct field

func (t Time) MarshalText() ([]byte, error) {
	if y := t.Year(); y < 0 || y >= 10000 {
		return nil, errors.New("Time.MarshalText: year outside of range [0,9999]")
	}

	b := make([]byte, 0, len(t.Format))
	return t.AppendFormat(b, t.Format), nil
}

func (t Time) MarshalJSON() (data []byte, err error) {
	if y := t.Year(); y < 0 || y >= 10000 {
		return nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")
	}

	b := make([]byte, 0, len(t.Format)+2)
	b = append(b, '"')
	b = t.AppendFormat(b, t.Format)
	b = append(b, '"')
	return b, nil
}

func (t *Time) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}
	// Fractional seconds are handled implicitly by Parse.
	var err error
	t.Time, err = time.Parse(`"`+t.Format+`"`, string(data))
	return err
}

func (t *Time) UnmarshalText(data []byte) error {
	// Fractional seconds are handled implicitly by Parse.
	var err error
	t.Time, err = time.Parse(t.Format, string(data))
	return err
}

// JSONString is used to denote this string should be treated as a JSON
type JSONString string

func (s JSONString) ToMap() any {
	var m any
	if s[0] == '[' && s[len(s)-1] == ']' {
		m = make([]any, 0)
	} else {
		m = make(map[string]any)
	}
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		// return a response with the error message
		return fmt.Sprintf("invalid JSON: %v %v", s, err)
	}
	return m
}

// SetName lets you define a readable name for the JSONString.
// This only needs to be called once
func (s JSONString) SetName(name string) JSONString {
	m, ok := s.ToMap().(map[string]any)
	if !ok {
		return s // not a map, cannot set name
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	SetSchemaName(name, keys)
	return s
}

// SetSchemaName sets a name for the hash16 values of the JSON keys provided.
// This is used for the JSONString type to create a unique schema name based on the keys of the JSON object.
func SetSchemaName(name string, keys []string) {
	sort.Strings(keys)
	key := hash16(strings.Join(keys, ""))
	if v, exists := namedSchemas[key]; exists {
		if v == name {
			return // already set to the same name, no need to log
		}
		log.Printf("%v overrides named schema: %v -> %v", key, v, name)
	}
	namedSchemas[key] = name
}

var namedSchemas = map[string]string{} // [hash16_key]name

// GetSchemaName returns a unique name for the schema based on the keys provided.
// it will use a predefined name if it exists
func GetSchemaName(keys []string) string {
	sort.Strings(keys)
	key := hash16(strings.Join(keys, ""))
	if name, exists := namedSchemas[key]; exists {
		return name
	}
	return key
}
