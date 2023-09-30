package typedcsv

import (
	"encoding"
	"encoding/csv"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// A TypedCSVWriter writes structs to a CSV file.
//
// The struct must have exported fields with a "csv" tag.
//
//   - the "csv" tag value is used as the CSV header.
//   - the "null" tag value is used as the CSV value when the field is nil.
//   - the "format" tag value is used as the CSV value. The format and the field value are passed to fmt.Sprintf.
//   - the "time_format" tag value is used to format time.Time fields. The value must be a valid time.Time format.
//   - the "time_location" tag value is used to set the location of time.Time fields. The value must be a valid time.Location name. Should be used with the "time_format" tag value.
//   - the "separator" tag value is used to join slice fields. Can be used with the "format" tag value.
//
// If a field implements encoding.TextMarshaler, the CSV value is the result of calling MarshalText.
type TypedCSVWriter[T any] struct {
	Writer *csv.Writer
}

// NewWriter returns a new TypedCSVWriter that wraps the given csv.Writer.
func NewWriter[T any](writer *csv.Writer) *TypedCSVWriter[T] {
	return &TypedCSVWriter[T]{
		Writer: writer,
	}
}

// WriteHeader writes the CSV header to the underlying writer.
// It uses the "csv" tag value of the struct fields.
func (w *TypedCSVWriter[T]) WriteHeader() error {
	var zero [0]T
	t := reflect.TypeOf(zero).Elem()

	var header []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if isValidCSVField(field) {
			header = append(header, field.Tag.Get(csvTag))
		}
	}

	return w.Writer.Write(header)
}

// WriteRecord writes the CSV record to the underlying writer.
// It returns a FieldFormatError if a field cannot be formatted.
// Otherwise, it returns any error returned by the underlying writer.
func (w *TypedCSVWriter[T]) WriteRecord(record T) error {
	recordType := reflect.TypeOf(record)
	recordValue := reflect.ValueOf(record)

	var values []string
	for i := 0; i < recordType.NumField(); i++ {
		field := recordType.Field(i)
		if !isValidCSVField(field) {
			continue
		}
		csvTagValue := field.Tag.Get(csvTag)
		fieldValue := recordValue.Field(i)
		fieldKind := fieldValue.Kind()
		// Pointer
		if fieldKind == reflect.Ptr {
			if fieldValue.IsNil() {
				nullTagValue := field.Tag.Get(nullTag)
				values = append(values, nullTagValue)
				continue
			}
			fieldValue = fieldValue.Elem()
		}
		fieldType := fieldValue.Type()
		// Time
		if fieldType.ConvertibleTo(timeType) {
			if timeFormat, ok := field.Tag.Lookup(timeFormatTag); ok {
				timeValue := fieldValue.Convert(timeType).Interface().(time.Time)
				if timeLocation, ok := field.Tag.Lookup(timeLocationTag); ok {
					location, err := time.LoadLocation(timeLocation)
					if err != nil {
						return FieldFormatError{Field: csvTagValue, NestedError: err}
					}

					timeValue = timeValue.In(location)
				}

				values = append(values, timeValue.Format(timeFormat))
				continue
			}
		}
		// TextMarshaler
		if fieldType.Implements(textMarshalerType) {
			text, err := fieldValue.Interface().(encoding.TextMarshaler).MarshalText()
			if err != nil {
				return FieldFormatError{Field: csvTagValue, NestedError: err}
			}
			values = append(values, string(text))
			continue
		}
		// Slice
		if fieldKind == reflect.Slice {
			separator := field.Tag.Get(separatorTag)
			format, ok := field.Tag.Lookup(formatTag)
			if !ok {
				format = "%v"
			}
			var builder strings.Builder
			for i := 0; i < fieldValue.Len(); i++ {
				if i > 0 {
					builder.WriteString(separator)
				}
				builder.WriteString(fmt.Sprintf(format, fieldValue.Index(i).Interface()))
			}
			values = append(values, builder.String())
			continue
		}
		// Format
		if format, ok := field.Tag.Lookup(formatTag); ok {
			values = append(values, fmt.Sprintf(format, fieldValue.Interface()))
			continue
		}
		// Default
		values = append(values, fmt.Sprintf("%v", fieldValue.Interface()))
	}

	return w.Writer.Write(values)
}

// Flush writes any buffered data to the underlying csv.Writer.
// To check if an error occurred during the Flush, call Error.
func (w *TypedCSVWriter[T]) Flush() {
	w.Writer.Flush()
}

// Error reports any error that has occurred during a previous WriteHeader, WriteRecord or Flush.
func (w *TypedCSVWriter[T]) Error() error {
	return w.Writer.Error()
}
