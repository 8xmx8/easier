package cvt

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToString(t *testing.T) {
	type caseT struct {
		name     string
		value    interface{}
		valueStr string
	}

	cases := []*caseT{
		{
			name:     "string",
			value:    "hello",
			valueStr: "hello",
		},
		{
			name:     "int",
			value:    1,
			valueStr: "1",
		},
		{
			name:     "float",
			value:    1.1,
			valueStr: "1.1",
		},
		{
			name:     "bool",
			value:    true,
			valueStr: "true",
		},
		{
			name:     "nil",
			value:    nil,
			valueStr: "",
		},
		{
			name:     "map",
			value:    map[string]string{"a": "b"},
			valueStr: "{\"a\":\"b\"}",
		},
		{
			name:     "slice",
			value:    []string{"a", "b"},
			valueStr: "[\"a\",\"b\"]",
		},
	}

	for _, cc := range cases {
		actual := ToString(cc.value)
		assert.Equal(t, cc.valueStr, actual)
	}
}

func TestToStringWithDefault(t *testing.T) {
	type args struct {
		value interface{}
		def   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "string",
			args: args{
				value: "hello",
				def:   "world",
			},
			want: "hello",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, ToStringWithDefault(tt.args.value, tt.args.def), "ToStringWithDefault(%v, %v)", tt.args.value, tt.args.def)
		})
	}
}
