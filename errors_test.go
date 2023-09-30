package typedcsv_test

import (
	"errors"
	"testing"
	"typedcsv"
)

func TestFieldParseError(t *testing.T) {
	customErr := errors.New("custom error")
	err := &typedcsv.FieldParseError{
		Field:       "custom field name",
		NestedError: customErr,
	}
	expected := "typedcsv: error parsing field 'custom field name': custom error"
	if err.Error() != expected {
		t.Fatalf("Expected %v, got %v", expected, err.Error())
	}
	if errors.Unwrap(err) != customErr {
		t.Fatalf("Expected %v, got %v", customErr, errors.Unwrap(err))
	}
}
