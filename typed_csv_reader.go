package typedcsv

import (
	"encoding"
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"
)

// A TypedCSVReader reads structs from a CSV file.
//
// The struct must have exported fields with a "csv" tag.
//
//   - the "csv" tag value is used as the CSV header.
//   - the "null" tag value is used to set the field to nil when the CSV value is equal to the tag value.
//   - the "time_format" tag value is used to parse time.Time fields. The value must be a valid time.Time format.
//   - the "time_location" tag value is used to set the location of time.Time fields. The value must be a valid time.Location name. Should be used with the "time_format" tag value.
//   - the "separator" tag value is used to split slice fields.
//
// If a field implements encoding.TextUnmarshaler, the CSV value is passed to UnmarshalText.
type TypedCSVReader[T any] struct {
	Reader *csv.Reader
	Header map[string]int
}

// NewReader returns a new TypedCSVReader that wraps the given csv.Reader.
func NewReader[T any](reader *csv.Reader) *TypedCSVReader[T] {
	return &TypedCSVReader[T]{
		Reader: reader,
	}
}

// ReadHeader reads the CSV header from the underlying reader.
// It uses the "csv" tag value of the struct fields.
// It returns io.EOF if there is no header.
func (r *TypedCSVReader[T]) ReadHeader() error {
	header, err := r.Reader.Read()
	if err != nil {
		return err
	}
	r.Header = make(map[string]int)
	for i, field := range header {
		r.Header[field] = i
	}
	return nil
}

// ReadRecord reads the CSV record from the underlying reader.
// It returns ErrHeaderNotRead if ReadHeader was not called.
// It returns io.EOF if there are no more records.
// It returns a FieldParseError if a field cannot be parsed.
// Otherwise, it returns any error returned by the underlying reader.
func (r *TypedCSVReader[T]) ReadRecord() (record *T, err error) {
	if r.Header == nil {
		err = ErrHeaderNotRead
		return
	}

	values, err := r.Reader.Read()
	if err != nil {
		return
	}

	record = new(T)

	recordType := reflect.TypeOf(record).Elem()
	recordValue := reflect.ValueOf(record).Elem()

	for i := 0; i < recordType.NumField(); i++ {
		field := recordType.Field(i)
		if !isValidCSVField(field) {
			continue
		}
		csvTagValue := field.Tag.Get(csvTag)
		index, ok := r.Header[csvTagValue]
		if !ok {
			continue
		}
		value := values[index]
		fieldValue := recordValue.Field(i)
		fieldKind := fieldValue.Kind()
		// Pointer
		if fieldKind == reflect.Ptr {
			if nullTagValue, ok := field.Tag.Lookup(nullTag); ok && value == nullTagValue {
				fieldValue.Set(reflect.Zero(fieldValue.Type()))
				continue
			}
			fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
			fieldValue = fieldValue.Elem()
		}
		fieldType := fieldValue.Type()
		fieldAddr := fieldValue.Addr()
		fieldAddrInterface := fieldAddr.Interface()
		// Time
		if fieldType.ConvertibleTo(timeType) {
			timeFormat := field.Tag.Get(timeFormatTag)
			var timeValue time.Time
			if timeFormat != "" {
				// time location tag
				timeLocation := field.Tag.Get(timeLocationTag)
				if timeLocation != "" {
					location, err := time.LoadLocation(timeLocation)
					if err != nil {
						return record, FieldParseError{Field: csvTagValue, NestedError: err}
					}
					timeValue, err = time.ParseInLocation(timeFormat, value, location)
					if err != nil {
						return record, FieldParseError{Field: csvTagValue, NestedError: err}
					}
				} else {
					timeValue, err = time.Parse(timeFormat, value)
					if err != nil {
						return record, FieldParseError{Field: csvTagValue, NestedError: err}
					}
				}
				fieldValue.Set(reflect.ValueOf(timeValue).Convert(fieldType))
				continue
			}
		}
		// TextUnmarshaler
		if fieldAddr.Type().Implements(textUnmarshalerType) {
			err := fieldAddrInterface.(encoding.TextUnmarshaler).UnmarshalText([]byte(value))
			if err != nil {
				return record, FieldParseError{Field: csvTagValue, NestedError: err}
			}
			continue
		}
		// Slice
		if fieldKind == reflect.Slice {
			separator := field.Tag.Get(separatorTag)
			slice := reflect.MakeSlice(fieldType, 0, 0)
			for itemIndex, item := range strings.Split(value, separator) {
				itemValue := reflect.New(fieldType.Elem())
				_, err := fmt.Sscanf(item, "%v", itemValue.Interface())
				if err != nil {
					return record, FieldParseError{Field: fmt.Sprintf("%s[%d]", csvTagValue, itemIndex), NestedError: err}
				}
				slice = reflect.Append(slice, itemValue.Elem())
			}
			fieldValue.Set(slice)
			continue
		}
		// Default
		_, err := fmt.Sscanf(value, "%v", fieldAddrInterface)
		if err == io.EOF {
			fieldValue.Set(reflect.Zero(fieldValue.Type()))
			err = nil
		}
		if err != nil {
			return record, FieldParseError{Field: csvTagValue, NestedError: err}
		}
	}

	return
}

// ReadAll reads all the remaining records from the underlying reader.
// It returns ErrHeaderNotRead if ReadHeader was not called.
// It returns a FieldParseError if a field cannot be parsed.
// Otherwise, it returns any error returned by the underlying reader.
func (r *TypedCSVReader[T]) ReadAll() (records []*T, err error) {
	for {
		record, err := r.ReadRecord()
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			return records, err
		}
		records = append(records, record)
	}
	return
}
