package utils

import (
	"reflect"
	"testing"
)

func Test_SplitEscapedString(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"a", args{"a"}, []string{"a"}},
		{"a b", args{"a b"}, []string{"a", "b"}},
		{`a\ b`, args{`a\ b`}, []string{"a b"}},
		{`a\ b c`, args{`a\ b c`}, []string{"a b", "c"}},
		{`a\ b c\ d`, args{`a\ b c\ d`}, []string{"a b", "c d"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SplitEscapedString(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitEscapedString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ExamplePrettyPrint() {
	type A struct {
		A int
		B string
	}
	a := A{1, "hello"}
	PrettyPrint(a)
	// Output:
	// {
	//   "A": 1,
	//   "B": "hello"
	// }
}
