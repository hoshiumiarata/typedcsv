package typedcsv_test

import (
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"reflect"
	"testing"
	"time"
	"typedcsv"
)

func TestReadHeader(t *testing.T) {
	reader := bytes.Buffer{}
	reader.WriteString("name,birthday,age,pet names,active,status,percentage,optional\n")
	csvReader := typedcsv.NewReader[Person](csv.NewReader(&reader))
	err := csvReader.ReadHeader()
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string]int{
		"name":       0,
		"birthday":   1,
		"age":        2,
		"pet names":  3,
		"active":     4,
		"status":     5,
		"percentage": 6,
		"optional":   7,
	}
	if !reflect.DeepEqual(csvReader.Header, expected) {
		t.Fatalf("Expected %v, got %v", expected, csvReader.Header)
	}
}

func TestReadHeaderEmpty(t *testing.T) {
	reader := bytes.Buffer{}
	csvReader := typedcsv.NewReader[Person](csv.NewReader(&reader))
	err := csvReader.ReadHeader()
	if err != io.EOF {
		t.Fatalf("Expected %v, got %v", io.EOF, err)
	}
}

func TestReadRecordMultiple(t *testing.T) {
	reader := bytes.Buffer{}
	reader.WriteString("name,birthday,age,pet names,active,status,percentage,optional\n")
	reader.WriteString("John,1970-06-17,55,Fluffy;Spot,true,active,12.35,NULL\n")
	reader.WriteString("Mary,1971-07-18,66,Puffy;Rover,false,inactive,23.46,Hello\n")
	csvReader := typedcsv.NewReader[Person](csv.NewReader(&reader))
	err := csvReader.ReadHeader()
	if err != nil {
		t.Fatal(err)
	}
	record, err := csvReader.ReadRecord()
	if err != nil {
		t.Fatal(err)
	}
	expected := &Person{
		Name:       "John",
		Birthday:   time.Date(1970, 6, 17, 0, 0, 0, 0, time.UTC),
		Age:        55,
		PetNames:   []string{"Fluffy", "Spot"},
		Active:     true,
		Status:     PersonStatusActive,
		Percentage: 12.35,
		Optional:   nil,
	}
	if !reflect.DeepEqual(record, expected) {
		t.Fatalf("Expected %v, got %v", expected, record)
	}

	record, err = csvReader.ReadRecord()
	if err != nil {
		t.Fatal(err)
	}
	str := "Hello"
	expected = &Person{
		Name:       "Mary",
		Birthday:   time.Date(1971, 7, 18, 0, 0, 0, 0, time.UTC),
		Age:        66,
		PetNames:   []string{"Puffy", "Rover"},
		Active:     false,
		Status:     PersonStatusInactive,
		Percentage: 23.46,
		Optional:   &str,
	}
	if !reflect.DeepEqual(record, expected) {
		t.Fatalf("Expected %v, got %v", expected, record)
	}

	_, err = csvReader.ReadRecord()
	if err != io.EOF {
		t.Fatalf("Expected %v, got %v", io.EOF, err)
	}
}

func TestReadRecordTime(t *testing.T) {
	reader := bytes.Buffer{}
	reader.WriteString("time,custom_time,time_without_format\n")
	reader.WriteString("1970-06-17 01:02:03,1971-07-18 02:03:04,1972-08-19T03:04:05+09:00\n")
	csvReader := typedcsv.NewReader[TimeTestRecord](csv.NewReader(&reader))
	err := csvReader.ReadHeader()
	if err != nil {
		t.Fatal(err)
	}
	record, err := csvReader.ReadRecord()
	if err != nil {
		t.Fatal(err)
	}
	expected := &TimeTestRecord{
		Time:              time.Date(1970, 6, 17, 1, 2, 3, 0, time.FixedZone("Asia/Tokyo", 9*60*60)),
		CustomTime:        CustomTime(time.Date(1971, 7, 18, 2, 3, 4, 0, time.FixedZone("Asia/Tokyo", 9*60*60))),
		TimeWithoutFormat: time.Date(1972, 8, 19, 3, 4, 5, 0, time.FixedZone("Asia/Tokyo", 9*60*60)),
	}

	if !record.Time.Equal(expected.Time) {
		t.Fatalf("Expected %v, got %v", expected.Time, record.Time)
	}
	if !time.Time(record.CustomTime).Equal(time.Time(expected.CustomTime)) {
		t.Fatalf("Expected %v, got %v", expected.CustomTime, record.CustomTime)
	}
	if !record.TimeWithoutFormat.Equal(expected.TimeWithoutFormat) {
		t.Fatalf("Expected %v, got %v", expected.TimeWithoutFormat, record.TimeWithoutFormat)
	}
}

func TestReadRecordTimeWithWrongTimeFormat(t *testing.T) {
	reader := bytes.Buffer{}
	reader.WriteString("time_with_location,time_without_location\n")
	reader.WriteString("abc,1970-06-17 01:02:03\n")
	csvReader := typedcsv.NewReader[TimeFormatTestRecord](csv.NewReader(&reader))
	err := csvReader.ReadHeader()
	if err != nil {
		t.Fatal(err)
	}
	_, err = csvReader.ReadRecord()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	var fieldParseError typedcsv.FieldParseError
	if !errors.As(err, &fieldParseError) {
		t.Fatalf("Expected %T, got %T", fieldParseError, err)
	}
	if fieldParseError.Field != "time_with_location" {
		t.Fatalf("Expected %v, got %v", "time_with_location", fieldParseError.Field)
	}
	expected := "parsing time \"abc\" as \"2006-01-02 15:04:05\": cannot parse \"abc\" as \"2006\""
	if fieldParseError.Unwrap().Error() != expected {
		t.Fatalf("Expected %v, got %v", expected, fieldParseError.Unwrap().Error())
	}
	expected = "typedcsv: error parsing field 'time_with_location': parsing time \"abc\" as \"2006-01-02 15:04:05\": cannot parse \"abc\" as \"2006\""
	if err.Error() != expected {
		t.Fatalf("Expected %v, got %v", expected, err.Error())
	}

	reader.Reset()
	reader.WriteString("time_with_location,time_without_location\n")
	reader.WriteString("1970-06-17 01:02:03,abc\n")
	csvReader = typedcsv.NewReader[TimeFormatTestRecord](csv.NewReader(&reader))
	err = csvReader.ReadHeader()
	if err != nil {
		t.Fatal(err)
	}
	_, err = csvReader.ReadRecord()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !errors.As(err, &fieldParseError) {
		t.Fatalf("Expected %T, got %T", fieldParseError, err)
	}
	if fieldParseError.Field != "time_without_location" {
		t.Fatalf("Expected %v, got %v", "time_without_location", fieldParseError.Field)
	}
	expected = "parsing time \"abc\" as \"2006-01-02 15:04:05\": cannot parse \"abc\" as \"2006\""
	if fieldParseError.Unwrap().Error() != expected {
		t.Fatalf("Expected %v, got %v", expected, fieldParseError.Unwrap().Error())
	}
	expected = "typedcsv: error parsing field 'time_without_location': parsing time \"abc\" as \"2006-01-02 15:04:05\": cannot parse \"abc\" as \"2006\""
	if err.Error() != expected {
		t.Fatalf("Expected %v, got %v", expected, err.Error())
	}
}

func TestReadRecordTimeWithWrongTimeLocation(t *testing.T) {
	reader := bytes.Buffer{}
	reader.WriteString("time\n")
	reader.WriteString("1970-06-17 01:02:03\n")
	csvReader := typedcsv.NewReader[TimeWithWrongTimeLocationTestRecord](csv.NewReader(&reader))
	err := csvReader.ReadHeader()
	if err != nil {
		t.Fatal(err)
	}
	_, err = csvReader.ReadRecord()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	var fieldParseError typedcsv.FieldParseError
	if !errors.As(err, &fieldParseError) {
		t.Fatalf("Expected %T, got %T", fieldParseError, err)
	}
	if fieldParseError.Field != "time" {
		t.Fatalf("Expected %v, got %v", "time", fieldParseError.Field)
	}
	expected := "unknown time zone abcdef"
	if fieldParseError.Unwrap().Error() != expected {
		t.Fatalf("Expected %v, got %v", expected, fieldParseError.Unwrap().Error())
	}
	expected = "typedcsv: error parsing field 'time': unknown time zone abcdef"
	if err.Error() != expected {
		t.Fatalf("Expected %v, got %v", expected, err.Error())
	}
}

func TestReadRecordOptional(t *testing.T) {
	reader := bytes.Buffer{}
	reader.WriteString("optional_string,optional_string_with_empty_tag,optional_time\n")
	reader.WriteString(",,NULL\n")
	csvReader := typedcsv.NewReader[OptionalTestRecord](csv.NewReader(&reader))
	err := csvReader.ReadHeader()
	if err != nil {
		t.Fatal(err)
	}
	record, err := csvReader.ReadRecord()
	if err != nil {
		t.Fatal(err)
	}
	str := ""
	expected := &OptionalTestRecord{
		OptionalStringWithoutTag:   &str,
		OptionalStringWithEmptyTag: nil,
		OptionalTime:               nil,
	}
	if !reflect.DeepEqual(record, expected) {
		t.Fatalf("Expected %v, got %v", expected, record)
	}
}

func TestReadRecordSlice(t *testing.T) {
	reader := bytes.Buffer{}
	reader.WriteString("slice,slice_with_new_line,slice_without_separator\n")
	reader.WriteString("a;b;c,\"d\ne\nf\",ghi\n")
	csvReader := typedcsv.NewReader[SliceTestRecord](csv.NewReader(&reader))
	err := csvReader.ReadHeader()
	if err != nil {
		t.Fatal(err)
	}
	record, err := csvReader.ReadRecord()
	if err != nil {
		t.Fatal(err)
	}
	expected := &SliceTestRecord{
		Slice:                 []string{"a", "b", "c"},
		SliceWithNewLine:      []string{"d", "e", "f"},
		SliceWithoutSeparator: []string{"g", "h", "i"},
	}
	if !reflect.DeepEqual(record, expected) {
		t.Fatalf("Expected %v, got %v", expected, record)
	}
}

func TestReadRecordUnmarshalText(t *testing.T) {
	reader := bytes.Buffer{}
	reader.WriteString("person_status\n")
	reader.WriteString("active\n")
	csvReader := typedcsv.NewReader[MarshalTextTestRecord](csv.NewReader(&reader))
	err := csvReader.ReadHeader()
	if err != nil {
		t.Fatal(err)
	}
	record, err := csvReader.ReadRecord()
	if err != nil {
		t.Fatal(err)
	}
	expected := &MarshalTextTestRecord{
		PersonStatus: PersonStatusActive,
	}
	if !reflect.DeepEqual(record, expected) {
		t.Fatalf("Expected %v, got %v", expected, record)
	}

	reader.Reset()
	reader.WriteString("person_status\n")
	reader.WriteString("abcdef\n")
	csvReader = typedcsv.NewReader[MarshalTextTestRecord](csv.NewReader(&reader))
	err = csvReader.ReadHeader()
	if err != nil {
		t.Fatal(err)
	}
	_, err = csvReader.ReadRecord()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	var fieldParseError typedcsv.FieldParseError
	if !errors.As(err, &fieldParseError) {
		t.Fatalf("Expected %T, got %T", fieldParseError, err)
	}
	if fieldParseError.Field != "person_status" {
		t.Fatalf("Expected %v, got %v", "person_status", fieldParseError.Field)
	}
	if fieldParseError.Unwrap().Error() != "unknown status" {
		t.Fatalf("Expected %v, got %v", "unknown status", fieldParseError.Unwrap().Error())
	}
	expectedErrorMessage := "typedcsv: error parsing field 'person_status': unknown status"
	if err.Error() != expectedErrorMessage {
		t.Fatalf("Expected %v, got %v", expectedErrorMessage, err.Error())
	}
}

func TestReadRecordWithoutReadingHeader(t *testing.T) {
	reader := bytes.Buffer{}
	reader.WriteString("name,birthday,age,pet names,active,status,percentage,optional\n")
	reader.WriteString("John,1970-06-17,55,Fluffy;Spot,true,active,12.35,NULL\n")
	csvReader := typedcsv.NewReader[Person](csv.NewReader(&reader))
	_, err := csvReader.ReadRecord()
	if err != typedcsv.ErrHeaderNotRead {
		t.Fatalf("Expected %v, got %v", typedcsv.ErrHeaderNotRead, err)
	}
}

func TestReadRecordEmpty(t *testing.T) {
	reader := bytes.Buffer{}
	reader.WriteString("header\n")
	csvReader := typedcsv.NewReader[Person](csv.NewReader(&reader))
	err := csvReader.ReadHeader()
	if err != nil {
		t.Fatal(err)
	}
	_, err = csvReader.ReadRecord()
	if err != io.EOF {
		t.Fatalf("Expected %v, got %v", io.EOF, err)
	}
}

func TestReadRecordNotExistingField(t *testing.T) {
	reader := bytes.Buffer{}
	reader.WriteString("header\n")
	reader.WriteString("value\n")
	csvReader := typedcsv.NewReader[Person](csv.NewReader(&reader))
	err := csvReader.ReadHeader()
	if err != nil {
		t.Fatal(err)
	}
	_, err = csvReader.ReadRecord()
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadRecordNotSupportedTypes(t *testing.T) {
	reader := bytes.Buffer{}
	reader.WriteString("map\n")
	reader.WriteString("1:2\n")
	csvReader1 := typedcsv.NewReader[MapTestRecord](csv.NewReader(&reader))
	err := csvReader1.ReadHeader()
	if err != nil {
		t.Fatal(err)
	}
	_, err = csvReader1.ReadRecord()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	var fieldParseError typedcsv.FieldParseError
	if !errors.As(err, &fieldParseError) {
		t.Fatalf("Expected %T, got %T", fieldParseError, err)
	}
	if fieldParseError.Field != "map" {
		t.Fatalf("Expected %v, got %v", "map", fieldParseError.Field)
	}
	expected := "can't scan type: *map[string]string"
	if fieldParseError.Unwrap().Error() != expected {
		t.Fatalf("Expected %v, got %v", expected, fieldParseError.Unwrap().Error())
	}
	expected = "typedcsv: error parsing field 'map': can't scan type: *map[string]string"
	if err.Error() != expected {
		t.Fatalf("Expected %v, got %v", expected, err.Error())
	}

	reader.Reset()
	reader.WriteString("slice_of_map\n")
	reader.WriteString("1:2\n")
	csvReader2 := typedcsv.NewReader[SliceOfMapTestRecord](csv.NewReader(&reader))
	err = csvReader2.ReadHeader()
	if err != nil {
		t.Fatal(err)
	}
	_, err = csvReader2.ReadRecord()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !errors.As(err, &fieldParseError) {
		t.Fatalf("Expected %T, got %T", fieldParseError, err)
	}
	if fieldParseError.Field != "slice_of_map[0]" {
		t.Fatalf("Expected %v, got %v", "slice_of_map[0]", fieldParseError.Field)
	}
	expected = "can't scan type: *map[string]string"
	if fieldParseError.Unwrap().Error() != expected {
		t.Fatalf("Expected %v, got %v", expected, fieldParseError.Unwrap().Error())
	}
	expected = "typedcsv: error parsing field 'slice_of_map[0]': can't scan type: *map[string]string"
	if err.Error() != expected {
		t.Fatalf("Expected %v, got %v", expected, err.Error())
	}
}

func TestReadAll(t *testing.T) {
	reader := bytes.Buffer{}
	reader.WriteString("name,birthday,age,pet names,active,status,percentage,optional\n")
	reader.WriteString("John,1970-06-17,55,Fluffy;Spot,true,active,12.35,NULL\n")
	reader.WriteString("Mary,1971-07-18,66,Puffy;Rover,false,inactive,23.46,NULL\n")
	csvReader := typedcsv.NewReader[Person](csv.NewReader(&reader))
	err := csvReader.ReadHeader()
	if err != nil {
		t.Fatal(err)
	}
	records, err := csvReader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	expected := []*Person{
		{
			Name:       "John",
			Birthday:   time.Date(1970, 6, 17, 0, 0, 0, 0, time.UTC),
			Age:        55,
			PetNames:   []string{"Fluffy", "Spot"},
			Active:     true,
			Status:     PersonStatusActive,
			Percentage: 12.35,
			Optional:   nil,
		},
		{
			Name:       "Mary",
			Birthday:   time.Date(1971, 7, 18, 0, 0, 0, 0, time.UTC),
			Age:        66,
			PetNames:   []string{"Puffy", "Rover"},
			Active:     false,
			Status:     PersonStatusInactive,
			Percentage: 23.46,
			Optional:   nil,
		},
	}
	if !reflect.DeepEqual(records, expected) {
		t.Fatalf("Expected %v, got %v", expected, records)
	}
}

func TestReadAllTimeWithWrongTimeLocation(t *testing.T) {
	reader := bytes.Buffer{}
	reader.WriteString("time\n")
	reader.WriteString("1970-06-17 01:02:03\n")
	csvReader := typedcsv.NewReader[TimeWithWrongTimeLocationTestRecord](csv.NewReader(&reader))
	err := csvReader.ReadHeader()
	if err != nil {
		t.Fatal(err)
	}
	_, err = csvReader.ReadAll()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	var fieldParseError typedcsv.FieldParseError
	if !errors.As(err, &fieldParseError) {
		t.Fatalf("Expected %T, got %T", fieldParseError, err)
	}
	if fieldParseError.Field != "time" {
		t.Fatalf("Expected %v, got %v", "time", fieldParseError.Field)
	}
	expected := "unknown time zone abcdef"
	if fieldParseError.Unwrap().Error() != expected {
		t.Fatalf("Expected %v, got %v", expected, fieldParseError.Unwrap().Error())
	}
	expected = "typedcsv: error parsing field 'time': unknown time zone abcdef"
	if err.Error() != expected {
		t.Fatalf("Expected %v, got %v", expected, err.Error())
	}
}
