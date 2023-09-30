package typedcsv

import (
	"errors"
	"fmt"
)

// ErrHeaderNotRead is returned when ReadRecord is called before ReadHeader.
var ErrHeaderNotRead = errors.New("typedcsv: header not read")

// FieldParseError is returned when a field cannot be parsed.
type FieldParseError struct {
	// Field is the name of the field that could not be parsed.
	Field string
	// NestedError is the error returned by the underlying parser.
	NestedError error
}

// Error returns the error message.
func (e FieldParseError) Error() string {
	return fmt.Sprintf("typedcsv: error parsing field '%s': %v", e.Field, e.NestedError)
}

// Unwrap returns the nested error.
func (e FieldParseError) Unwrap() error {
	return e.NestedError
}

// FieldFormatError is returned when a field cannot be formatted.
type FieldFormatError struct {
	Field       string
	NestedError error
}

// Error returns the error message.
func (e FieldFormatError) Error() string {
	return fmt.Sprintf("typedcsv: error formatting field '%s': %v", e.Field, e.NestedError)
}

// Unwrap returns the nested error.
func (e FieldFormatError) Unwrap() error {
	return e.NestedError
}
