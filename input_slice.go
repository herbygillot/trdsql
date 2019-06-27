package trdsql

import (
	"fmt"
	"io"
	"reflect"
)

// SliceReader is a structure for reading tabular data in memory.
// It can be used as the trdsql reader interface.
type SliceReader struct {
	tableName string
	names     []string
	types     []string
	data      [][]interface{}
}

// NewSliceReader takes a tableName and tabular data in memory
// and returns SliceReader.
// The tabular data that can be received is
// a one-dimensional array,
// a two-dimensional array,
// a map,
// and an array of structures.
func NewSliceReader(tableName string, args interface{}) *SliceReader {
	val := reflect.ValueOf(args)

	switch val.Kind() {
	case reflect.Map:
		// {"1":"test"}
		return mapSliceReader(tableName, val)
	case reflect.Slice:
		// slice continue
	default:
		single := val.Interface()
		data := [][]interface{}{
			{single},
		}
		names := []string{"c1"}
		types := []string{typeToDBType(val.Kind())}
		return &SliceReader{
			tableName: tableName,
			names:     names,
			types:     types,
			data:      data,
		}
	}

	//
	switch val.Index(0).Kind() {
	case reflect.Struct:
		// {{ id: 1, name: "test"}}
		return structSliceReader(tableName, val)
	case reflect.Slice:
		// {{1, "test"}}
		return sliceSliceReader(tableName, val)
	default:
		// {{"a", "b", "c"}}
		return interfaceSliceReader(tableName, val)
	}
}

func mapSliceReader(tableName string, val reflect.Value) *SliceReader {
	val = reflect.Indirect(val)
	names := []string{"c1", "c2"}
	keyType := val.MapKeys()[0].Kind()
	valType := val.MapIndex(val.MapKeys()[0]).Kind()
	types := []string{typeToDBType(keyType), typeToDBType(valType)}
	data := make([][]interface{}, 0)
	for _, e := range val.MapKeys() {
		data = append(data, []interface{}{e.Interface(), val.MapIndex(e).Interface()})
	}
	return &SliceReader{
		tableName: tableName,
		names:     names,
		types:     types,
		data:      data,
	}
}

func interfaceSliceReader(tableName string, val reflect.Value) *SliceReader {
	single := val.Interface().([]interface{})
	names := []string{"c1"}
	t := reflect.ValueOf(single[0])
	types := []string{typeToDBType(t.Kind())}
	data := make([][]interface{}, 0)
	for i := 0; i < len(single); i++ {
		data = append(data, []interface{}{single[i]})
	}
	return &SliceReader{
		tableName: tableName,
		names:     names,
		types:     types,
		data:      data,
	}
}

func structSliceReader(tableName string, val reflect.Value) *SliceReader {
	length := val.Len()
	t := val.Index(0).Type()
	columnNum := t.NumField()
	names := make([]string, columnNum)
	types := make([]string, columnNum)
	for i := 0; i < columnNum; i++ {
		f := t.Field(i)
		names[i] = f.Name
		types[i] = typeToDBType(f.Type.Kind())
	}
	data := make([][]interface{}, 0)
	for i := 0; i < length; i++ {
		rows := val.Index(i)
		r := make([]interface{}, rows.NumField())
		for j := 0; j < rows.NumField(); j++ {
			r[j] = fmt.Sprintf("%v", rows.Field(j))
		}
		data = append(data, r)
	}
	return &SliceReader{
		tableName: tableName,
		names:     names,
		types:     types,
		data:      data,
	}
}

func sliceSliceReader(tableName string, val reflect.Value) *SliceReader {
	length := val.Len()
	col := val.Index(0)
	columnNum := col.Len()
	names := make([]string, columnNum)
	types := make([]string, columnNum)
	for i := 0; i < columnNum; i++ {
		names[i] = fmt.Sprintf("c%d", i+1)
		colType := reflect.ValueOf(col.Index(i).Interface()).Kind()
		types[i] = typeToDBType(colType)
	}

	data := make([][]interface{}, 0)
	for i := 0; i < length; i++ {
		data = append(data, val.Index(i).Interface().([]interface{}))
	}
	return &SliceReader{
		tableName: tableName,
		names:     names,
		types:     types,
		data:      data,
	}
}

// In sliceReader, only int type is passed to the database as int type.
func typeToDBType(t reflect.Kind) string {
	switch t {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "int"
	default:
		return DefaultDBType
	}
}

// Names returns column names.
func (r *SliceReader) Names() ([]string, error) {
	return r.names, nil
}

// Types returns column types.
func (r *SliceReader) Types() ([]string, error) {
	return r.types, nil
}

// PreReadRow is returns entity of the data.
func (r *SliceReader) PreReadRow() [][]interface{} {
	return r.data
}

// ReadRow only returns EOF.
func (r *SliceReader) ReadRow(row []interface{}) ([]interface{}, error) {
	return nil, io.EOF
}