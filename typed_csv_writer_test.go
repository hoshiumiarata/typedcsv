package typedcsv_test

import (
	"bytes"
	"encoding/csv"
	"errors"
	"testing"
	"time"
	"typedcsv"
)

func TestWriteHeader(t *testing.T) {
	writer := bytes.Buffer{}
	csvWriter := typedcsv.NewWriter[Person](csv.NewWriter(&writer))
	err := csvWriter.WriteHeader()
	if err != nil {
		t.Fatal(err)
	}
	csvWriter.Flush()
	expected := "name,birthday,age,pet names,active,status,percentage,optional\n"
	if writer.String() != expected {
		t.Fatalf("Expected %q, got %q", expected, writer.String())
	}
}

func TestWriteRecordMultiple(t *testing.T) {
	writer := bytes.Buffer{}
	csvWriter := typedcsv.NewWriter[Person](csv.NewWriter(&writer))
	err := csvWriter.WriteRecord(Person{
		Name:       "John",
		Birthday:   time.Date(1970, 6, 17, 0, 0, 0, 0, time.UTC),
		Age:        55,
		PetNames:   []string{"Fluffy", "Spot"},
		Active:     true,
		Status:     PersonStatusActive,
		Percentage: 12.3456,
		Optional:   nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	str1 := "Hello"
	err = csvWriter.WriteRecord(Person{
		Name:       "Mary",
		Birthday:   time.Date(1971, 7, 18, 0, 0, 0, 0, time.UTC),
		Age:        66,
		PetNames:   []string{"Puffy", "Rover"},
		Active:     false,
		Status:     PersonStatusInactive,
		Percentage: 23.4567,
		Optional:   &str1,
	})
	if err != nil {
		t.Fatal(err)
	}
	csvWriter.Flush()
	expected := "John,1970-06-17,55,Fluffy;Spot,true,active,12.35,NULL\nMary,1971-07-18,66,Puffy;Rover,false,inactive,23.46,Hello\n"
	if writer.String() != expected {
		t.Fatalf("Expected %q, got %q", expected, writer.String())
	}
}

func TestWriteRecordTime(t *testing.T) {
	writer := bytes.Buffer{}
	csvWriter := typedcsv.NewWriter[TimeTestRecord](csv.NewWriter(&writer))
	err := csvWriter.WriteRecord(TimeTestRecord{
		Time:              time.Date(1970, 6, 17, 1, 2, 3, 0, time.FixedZone("Asia/Tokyo", 9*60*60)),
		CustomTime:        CustomTime(time.Date(1971, 7, 18, 2, 3, 4, 0, time.FixedZone("Asia/Tokyo", 9*60*60))),
		TimeWithoutFormat: time.Date(1972, 8, 19, 3, 4, 5, 0, time.FixedZone("Asia/Tokyo", 9*60*60)),
	})
	if err != nil {
		t.Fatal(err)
	}
	csvWriter.Flush()
	expected := "1970-06-17 01:02:03,1971-07-18 02:03:04,1972-08-19T03:04:05+09:00\n"
	if writer.String() != expected {
		t.Fatalf("Expected %q, got %q", expected, writer.String())
	}
}

func TestWriteRecordTimeWithWrongTimeLocation(t *testing.T) {
	writer := bytes.Buffer{}
	csvWriter := typedcsv.NewWriter[TimeWithWrongTimeLocationTestRecord](csv.NewWriter(&writer))
	err := csvWriter.WriteRecord(TimeWithWrongTimeLocationTestRecord{
		Time: time.Date(1970, 6, 17, 1, 2, 3, 0, time.FixedZone("Asia/Tokyo", 9*60*60)),
	})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	var fieldFormatError typedcsv.FieldFormatError
	if !errors.As(err, &fieldFormatError) {
		t.Fatalf("Expected %T, got %T", fieldFormatError, err)
	}
	if fieldFormatError.Field != "time" {
		t.Fatalf("Expected %q, got %q", "time", fieldFormatError.Field)
	}
	expected := "unknown time zone abcdef"
	if fieldFormatError.Unwrap().Error() != expected {
		t.Fatalf("Expected %q, got %q", expected, fieldFormatError.NestedError.Error())
	}
	expected = "typedcsv: error formatting field 'time': unknown time zone abcdef"
	if err.Error() != expected {
		t.Fatalf("Expected %q, got %q", expected, err.Error())
	}
}

func TestWriteRecordOptional(t *testing.T) {
	writer := bytes.Buffer{}
	csvWriter := typedcsv.NewWriter[OptionalTestRecord](csv.NewWriter(&writer))
	err := csvWriter.WriteRecord(OptionalTestRecord{
		OptionalStringWithoutTag:   nil,
		OptionalStringWithEmptyTag: nil,
		OptionalTime:               nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	csvWriter.Flush()
	expected := ",,NULL\n"
	if writer.String() != expected {
		t.Fatalf("Expected %q, got %q", expected, writer.String())
	}

	writer.Reset()
	str1 := "Hello"
	str2 := "World"
	time := time.Date(1970, 6, 17, 1, 2, 3, 0, time.FixedZone("Asia/Tokyo", 9*60*60))
	err = csvWriter.WriteRecord(OptionalTestRecord{
		OptionalStringWithoutTag:   &str1,
		OptionalStringWithEmptyTag: &str2,
		OptionalTime:               &time,
	})
	if err != nil {
		t.Fatal(err)
	}
	csvWriter.Flush()
	expected = "Hello,World,1970-06-17T01:02:03+09:00\n"
	if writer.String() != expected {
		t.Fatalf("Expected %q, got %q", expected, writer.String())
	}
}

func TestWriteRecordSlice(t *testing.T) {
	writer := bytes.Buffer{}
	csvWriter := typedcsv.NewWriter[SliceTestRecord](csv.NewWriter(&writer))
	err := csvWriter.WriteRecord(SliceTestRecord{
		Slice:                 []string{"a", "b", "c"},
		SliceWithNewLine:      []string{"a", "b", "c"},
		SliceWithoutSeparator: []string{"a", "b", "c"},
	})
	if err != nil {
		t.Fatal(err)
	}
	csvWriter.Flush()
	expected := "a;b;c,\"a\nb\nc\",abc\n"
	if writer.String() != expected {
		t.Fatalf("Expected %q, got %q", expected, writer.String())
	}
}

func TestWriteRecordMarshalText(t *testing.T) {
	writer := bytes.Buffer{}
	csvWriter := typedcsv.NewWriter[MarshalTextTestRecord](csv.NewWriter(&writer))
	err := csvWriter.WriteRecord(MarshalTextTestRecord{
		PersonStatus: PersonStatusActive,
	})
	if err != nil {
		t.Fatal(err)
	}
	csvWriter.Flush()
	expected := "active\n"
	if writer.String() != expected {
		t.Fatalf("Expected %q, got %q", expected, writer.String())
	}

	writer.Reset()
	err = csvWriter.WriteRecord(MarshalTextTestRecord{
		PersonStatus: 100,
	})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	var fieldFormatError typedcsv.FieldFormatError
	if !errors.As(err, &fieldFormatError) {
		t.Fatalf("Expected %T, got %T", fieldFormatError, err)
	}
	if fieldFormatError.Field != "person_status" {
		t.Fatalf("Expected %q, got %q", "person_status", fieldFormatError.Field)
	}
	expected = "unknown status"
	if fieldFormatError.Unwrap().Error() != expected {
		t.Fatalf("Expected %q, got %q", expected, fieldFormatError.NestedError.Error())
	}
	expected = "typedcsv: error formatting field 'person_status': unknown status"
	if err.Error() != expected {
		t.Fatalf("Expected %q, got %q", expected, err.Error())
	}
}

func TestWriteRecordFormat(t *testing.T) {
	writer := bytes.Buffer{}
	csvWriter := typedcsv.NewWriter[FormatTestRecord](csv.NewWriter(&writer))
	err := csvWriter.WriteRecord(FormatTestRecord{
		Percentage: 12.3456,
		HexSlice:   []uint8{0x01, 0x02, 0x03},
	})
	if err != nil {
		t.Fatal(err)
	}
	csvWriter.Flush()
	expected := "12.35,010203\n"
	if writer.String() != expected {
		t.Fatalf("Expected %q, got %q", expected, writer.String())
	}
}

func TestWriterError(t *testing.T) {
	writer := &ErrorWriter{}
	csvWriter := typedcsv.NewWriter[Person](csv.NewWriter(writer))
	err := csvWriter.WriteHeader()
	if err != nil {
		t.Fatal(err)
	}
	csvWriter.Flush()
	err = csvWriter.Error()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}
