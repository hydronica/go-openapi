package openapi

import (
	"errors"
	"time"
)

type Time struct {
	time.Time
	Format string
}

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
