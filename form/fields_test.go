package form

import (
	"fmt"
	"reflect"
	"testing"
)

func TestFields(t *testing.T) {
	tests := map[string]struct {
		strct interface{}
		want  []field
	}{
		"Simplest use case": {
			strct: struct {
				Name string
			}{},
			want: []field{
				{
					Label:       "Name",
					Name:        "Name",
					Type:        "text",
					Placeholder: "Name",
					Value:       "",
				},
			},
		},
		"Field names should be determined from the struct": {
			strct: struct {
				FullName string
			}{},
			want: []field{
				{
					Label:       "FullName",
					Name:        "FullName",
					Type:        "text",
					Placeholder: "FullName",
					Value:       "",
				},
			},
		},
		"Multiple fields should be supported": {
			strct: struct {
				Name  string
				Email string
				Age   int
			}{},
			want: []field{
				{
					Label:       "Name",
					Name:        "Name",
					Type:        "text",
					Placeholder: "Name",
					Value:       "",
				},
				{
					Label:       "Email",
					Name:        "Email",
					Type:        "text",
					Placeholder: "Email",
					Value:       "",
				},
				{
					Label:       "Age",
					Name:        "Age",
					Type:        "text",
					Placeholder: "Age",
					Value:       0,
				},
			},
		},
		"Values should be parsed": {
			strct: struct {
				Name  string
				Email string
				Age   int
			}{
				Name:  "Jon Calhoun",
				Email: "jon@calhoun.io",
				Age:   123,
			},
			want: []field{
				{
					Label:       "Name",
					Name:        "Name",
					Type:        "text",
					Placeholder: "Name",
					Value:       "Jon Calhoun",
				},
				{
					Label:       "Email",
					Name:        "Email",
					Type:        "text",
					Placeholder: "Email",
					Value:       "jon@calhoun.io",
				},
				{
					Label:       "Age",
					Name:        "Age",
					Type:        "text",
					Placeholder: "Age",
					Value:       123,
				},
			},
		},
		"Unexported fields should be skipped": {
			strct: struct {
				Name  string
				email string
				Age   int
			}{
				Name:  "Jon Calhoun",
				email: "jon@calhoun.io",
				Age:   123,
			},
			want: []field{
				{
					Label:       "Name",
					Name:        "Name",
					Type:        "text",
					Placeholder: "Name",
					Value:       "Jon Calhoun",
				},
				{
					Label:       "Age",
					Name:        "Age",
					Type:        "text",
					Placeholder: "Age",
					Value:       123,
				},
			},
		},
		// "Non-structs shouldn't work": {
		// 	strct: "some string",
		// },
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := fields(tc.strct)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("fields() = %v; want %v", got, tc.want)
			}
		})
	}
}

func TestFields_invalidTypes(t *testing.T) {
	tests := []struct {
		notAStruct interface{}
	}{
		{"this is a string"},
		{123},
		{nil},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%T", tc.notAStruct), func(t *testing.T) {
			defer func() {
				if err := recover(); err == nil {
					t.Errorf("fields(%v) did not panic", tc.notAStruct)
				}
			}()
			fields(tc.notAStruct)
		})
	}
}

// func TestFields_labels(t *testing.T) {
// 	hasLabels := func(labels ...string) func(*testing.T, []field) {
// 		return func(t *testing.T, fields []field) {
// 			if len(fields) != len(labels) {
// 				t.Fatalf("fields() len = %d; want %d", len(fields), len(labels))
// 			}
// 			for i := 0; i < len(fields); i++ {
// 				if fields[i].Label != labels[i] {
// 					t.Errorf("fields()[%d].Label = %s; want %s", i, fields[i].Label, labels[i])
// 				}
// 			}
// 		}
// 	}
// 	hasValues := func(values ...interface{}) func(*testing.T, []field) {
// 		return func(t *testing.T, fields []field) {
// 			if len(fields) != len(values) {
// 				t.Fatalf("fields() len = %d; want %d", len(fields), len(values))
// 			}
// 			for i := 0; i < len(fields); i++ {
// 				if fields[i].Value != values[i] {
// 					t.Errorf("fields()[%d].Value = %v; want %v", i, fields[i].Value, values[i])
// 				}
// 			}
// 		}
// 	}
// 	check := func(checks ...func(*testing.T, []field)) []func(*testing.T, []field) {
// 		return checks
// 	}

// 	tests := map[string]struct {
// 		strct  interface{}
// 		checks []func(*testing.T, []field)
// 	}{
// 		"No values": {
// 			strct: struct {
// 				Name string
// 			}{},
// 			checks: check(hasLabels("Name")),
// 		},
// 		"Multiple fields with values": {
// 			strct: struct {
// 				Name  string
// 				Email string
// 				Age   int
// 			}{
// 				Name:  "Jon Calhoun",
// 				Email: "jon@calhoun.io",
// 				Age:   123,
// 			},
// 			checks: check(
// 				hasLabels("Name", "Email", "Age"),
// 				hasValues("Jon Calhoun", "jon@calhoun.io", 123),
// 			),
// 		},
// 	}
// 	for name, tc := range tests {
// 		t.Run(name, func(t *testing.T) {
// 			got := fields(tc.strct)
// 			for _, check := range tc.checks {
// 				check(t, got)
// 			}
// 		})
// 	}
// }
