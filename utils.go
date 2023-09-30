package typedcsv

import (
	"encoding"
	"reflect"
	"time"
)

const (
	csvTag          = "csv"
	nullTag         = "null"
	formatTag       = "format"
	timeFormatTag   = "time_format"
	timeLocationTag = "time_location"
	separatorTag    = "separator"
)

var (
	timeType            = reflect.TypeOf(time.Time{})
	textMarshalerType   = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

func isValidCSVField(field reflect.StructField) bool {
	return field.IsExported() && field.Tag.Get(csvTag) != ""
}
