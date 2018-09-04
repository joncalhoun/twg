package underscore

import (
	"fmt"
	"testing"
)

// func TestCamel(t *testing.T) {
// 	testCases := []struct {
// 		arg  string
// 		want string
// 	}{
// 		{"thisIsACamelCaseString", "this_is_a_camel_case_string"},
// 		{"with a space", "with a space"},
// 		{"endsWithA", "ends_with_a"},
// 	}
// 	for _, tc := range testCases {
// 		t.Logf("Testing: %q", tc.arg)
// 		got := Camel(tc.arg)
// 		if got != tc.want {
// 			t.Errorf("Camel(%q) = %q; want %q", tc.arg, got, tc.want)
// 		}
// 	}
// }

func TestCamel(t *testing.T) {
	tests := []struct {
		arg  string
		want string
	}{
		{"thisIsACamelCaseString", "this_is_a_camel_case_string1"},
		{"with a space", "with a space"},
		{"endsWithA", "ends_with_a1"},
	}
	// setup
	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			if got := Camel(tt.arg); got != tt.want {
				t.Fatalf("Camel() = %v, want %v", got, tt.want)
			}
			fmt.Println("this won't print if it fails...")
			// check2
			// check3
		})
	}
}
