package dialect

import (
	"reflect"
	"testing"
)

func TestMysql_DataTypeOf(t *testing.T) {
	s := &mysql{}
	cases := []struct {
		Value interface{}
		Type  string
	}{
		{"Tom", "varchar(20)"},
		{123, "int"},
		{1.2, "double"},
		{[]int{1, 2, 3}, "blob"},
	}

	for _, c := range cases {
		if typ := s.DataTypeOf(reflect.ValueOf(c.Value)); typ != c.Type {
			t.Fatalf("expect %s, but got %s", c.Type, typ)
		}
	}
}
