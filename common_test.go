package typedcsv_test

import (
	"errors"
	"time"
)

type Person struct {
	Name       string       `csv:"name"`
	Birthday   time.Time    `csv:"birthday" time_format:"2006-01-02"`
	Age        uint8        `csv:"age"`
	PetNames   []string     `csv:"pet names" separator:";"`
	Active     bool         `csv:"active"`
	Status     PersonStatus `csv:"status"`
	Percentage float64      `csv:"percentage" format:"%.2f"`
	Optional   *string      `csv:"optional" null:"NULL"`

	Skipped string
	_       bool
}

type PersonStatus uint8

const (
	PersonStatusUnknown PersonStatus = iota
	PersonStatusActive
	PersonStatusInactive
)

func (status PersonStatus) MarshalText() ([]byte, error) {
	switch status {
	case PersonStatusUnknown:
		return []byte("unknown"), nil
	case PersonStatusActive:
		return []byte("active"), nil
	case PersonStatusInactive:
		return []byte("inactive"), nil
	default:
		return nil, errors.New("unknown status")
	}
}

func (status *PersonStatus) UnmarshalText(text []byte) error {
	switch string(text) {
	case "unknown":
		*status = PersonStatusUnknown
	case "active":
		*status = PersonStatusActive
	case "inactive":
		*status = PersonStatusInactive
	default:
		return errors.New("unknown status")
	}
	return nil
}

type CustomTime time.Time

type TimeTestRecord struct {
	Time              time.Time  `csv:"time" time_format:"2006-01-02 15:04:05" time_location:"Asia/Tokyo"`
	CustomTime        CustomTime `csv:"custom_time" time_format:"2006-01-02 15:04:05" time_location:"Asia/Tokyo"`
	TimeWithoutFormat time.Time  `csv:"time_without_format"`
}

type TimeWithWrongTimeLocationTestRecord struct {
	Time time.Time `csv:"time" time_format:"2006-01-02 15:04:05" time_location:"abcdef"`
}

type TimeFormatTestRecord struct {
	TimeWithLocation    time.Time `csv:"time_with_location" time_format:"2006-01-02 15:04:05" time_location:"Asia/Tokyo"`
	TimeWithoutLocation time.Time `csv:"time_without_location" time_format:"2006-01-02 15:04:05"`
}

type OptionalTestRecord struct {
	OptionalStringWithoutTag   *string    `csv:"optional_string"`
	OptionalStringWithEmptyTag *string    `csv:"optional_string_with_empty_tag" null:""`
	OptionalTime               *time.Time `csv:"optional_time" null:"NULL"`
}

type SliceTestRecord struct {
	Slice                 []string `csv:"slice" separator:";"`
	SliceWithNewLine      []string `csv:"slice_with_new_line" separator:"\n"`
	SliceWithoutSeparator []string `csv:"slice_without_separator"`
}

type MarshalTextTestRecord struct {
	PersonStatus PersonStatus `csv:"person_status"`
}

type FormatTestRecord struct {
	Percentage float64 `csv:"percentage" format:"%.2f"`
	HexSlice   []uint8 `csv:"hex" format:"%02x" separator:""`
}

type MapTestRecord struct {
	Map map[string]string `csv:"map"`
}

type SliceOfMapTestRecord struct {
	SliceOfMap []map[string]string `csv:"slice_of_map"`
}

type ErrorWriter struct{}

func (w *ErrorWriter) Write([]byte) (int, error) {
	return 0, errors.New("write error")
}
