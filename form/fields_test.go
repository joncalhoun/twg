package form

import (
	"fmt"
	"reflect"
	"testing"
)

// TODO: Add test case for invalid struct tag value

func TestParseTags(t *testing.T) {
	tests := map[string]struct {
		arg  reflect.StructField
		want map[string]string
	}{
		"empty tag": {
			arg:  reflect.StructField{},
			want: nil,
		},
		"label tag": {
			arg: reflect.StructField{
				Tag: `form:"label=Full Name"`,
			},
			want: map[string]string{
				"label": "Full Name",
			},
		},
		"multiple tags": {
			arg: reflect.StructField{
				Tag: `form:"label=Full Name;name=full_name"`,
			},
			want: map[string]string{
				"label": "Full Name",
				"name":  "full_name",
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := parseTags(tc.arg)
			if len(got) != len(tc.want) {
				t.Errorf("parseTags() len = %d, want %d", len(got), len(tc.want))
			}
			for k, v := range tc.want {
				gv, ok := got[k]
				if !ok {
					t.Errorf("parseTags() missing key %q", k)
					continue
				}
				if gv != v {
					t.Errorf("parseTags()[%q] = %q; want %q", k, gv, v)
				}
				delete(got, k)
			}
			for gk, gv := range got {
				t.Errorf("parseTags() extra key %q, value = %q", gk, gv)
			}
		})
	}
}

func TestParseTags_invalidTag(t *testing.T) {
	tests := []struct {
		arg reflect.StructField
	}{
		{reflect.StructField{Tag: `form:"invalid-value"`}},
	}
	for _, tc := range tests {
		t.Run(string(tc.arg.Tag), func(t *testing.T) {
			defer func() {
				if err := recover(); err == nil {
					t.Errorf("parseTags() did not panic")
				}
			}()
			parseTags(tc.arg)
		})
	}
}

func TestFields(t *testing.T) {
	var nilStructPtr *struct {
		Name string
		Age  int
	}

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
		"Pointers to structs should be supported": {
			strct: &struct {
				Name string
				Age  int
			}{
				Name: "Jon Calhoun",
				Age:  123,
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
		"Nil pointers with a struct type should be supported": {
			strct: nilStructPtr,
			want: []field{
				{
					Label:       "Name",
					Name:        "Name",
					Type:        "text",
					Placeholder: "Name",
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
		"Pointer fields should be supported": {
			strct: struct {
				Name *string
				Age  *int
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
					Label:       "Age",
					Name:        "Age",
					Type:        "text",
					Placeholder: "Age",
					Value:       0,
				},
			},
		},
		"Nested structs should be supported": {
			strct: struct {
				Name    string
				Address struct {
					Street string
					Zip    int
				}
			}{
				Name: "Jon Calhoun",
				Address: struct {
					Street string
					Zip    int
				}{
					Street: "123 Fake St",
					Zip:    90210,
				},
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
					Label:       "Street",
					Name:        "Address.Street",
					Type:        "text",
					Placeholder: "Street",
					Value:       "123 Fake St",
				},
				{
					Label:       "Zip",
					Name:        "Address.Zip",
					Type:        "text",
					Placeholder: "Zip",
					Value:       90210,
				},
			},
		},
		"Doubly nested structs should be supported": {
			strct: struct {
				A struct {
					B struct {
						C1 string
						C2 int
					}
				}
			}{
				A: struct {
					B struct {
						C1 string
						C2 int
					}
				}{
					B: struct {
						C1 string
						C2 int
					}{
						C1: "C1-value",
						C2: 123,
					},
				},
			},
			want: []field{
				{
					Label:       "C1",
					Name:        "A.B.C1",
					Type:        "text",
					Placeholder: "C1",
					Value:       "C1-value",
				},
				{
					Label:       "C2",
					Name:        "A.B.C2",
					Type:        "text",
					Placeholder: "C2",
					Value:       123,
				},
			},
		},
		"Nested pointer structs should be supported": {
			strct: struct {
				Name    string
				Address *struct {
					Street string
					Zip    int
				}
				ContactCard *struct {
					Phone string
				}
			}{
				Name: "Jon Calhoun",
				Address: &struct {
					Street string
					Zip    int
				}{
					Street: "123 Fake St",
					Zip:    90210,
				},
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
					Label:       "Street",
					Name:        "Address.Street",
					Type:        "text",
					Placeholder: "Street",
					Value:       "123 Fake St",
				},
				{
					Label:       "Zip",
					Name:        "Address.Zip",
					Type:        "text",
					Placeholder: "Zip",
					Value:       90210,
				},
				{
					Label:       "Phone",
					Name:        "ContactCard.Phone",
					Type:        "text",
					Placeholder: "Phone",
					Value:       "",
				},
			},
		},
		"Struct tags": {
			strct: struct {
				LabelTest       string `form:"label=This is custom"`
				NameTest        string `form:"name=full_name"`
				TypeTest        int    `form:"type=number"`
				PlaceholderTest string `form:"placeholder=your value goes here..."`
				Nested          struct {
					MultiTest string `form:"name=NestedMulti;label=This is nested;type=email;placeholder=user@example.com"`
				}
			}{
				PlaceholderTest: "value and placeholder",
			},
			want: []field{
				{
					Label:       "This is custom",
					Name:        "LabelTest",
					Type:        "text",
					Placeholder: "LabelTest",
					Value:       "",
				},
				{
					Label:       "NameTest",
					Name:        "full_name",
					Type:        "text",
					Placeholder: "NameTest",
					Value:       "",
				},
				{
					Label:       "TypeTest",
					Name:        "TypeTest",
					Type:        "number",
					Placeholder: "TypeTest",
					Value:       0,
				},
				{
					Label:       "PlaceholderTest",
					Name:        "PlaceholderTest",
					Type:        "text",
					Placeholder: "your value goes here...",
					Value:       "value and placeholder",
				},
				{
					Label:       "This is nested",
					Name:        "NestedMulti",
					Type:        "email",
					Placeholder: "user@example.com",
					Value:       "",
				},
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := fields(tc.strct)
			if reflect.DeepEqual(got, tc.want) {
				return
			}
			if len(got) != len(tc.want) {
				t.Errorf("fields() len = %d; want %d", len(got), len(tc.want))
			}
			for i, gotField := range got {
				if i >= len(tc.want) {
					break
				}
				wantField := tc.want[i]
				if reflect.DeepEqual(gotField, wantField) {
					continue
				}
				t.Errorf("fields()[%d]", i)
				if gotField.Label != wantField.Label {
					t.Errorf("  .Label = %v; want %v", gotField.Label, wantField.Label)
				}
				if gotField.Name != wantField.Name {
					t.Errorf("  .Name = %v; want %v", gotField.Name, wantField.Name)
				}
				if gotField.Type != wantField.Type {
					t.Errorf("  .Type = %v; want %v", gotField.Type, wantField.Type)
				}
				if gotField.Placeholder != wantField.Placeholder {
					t.Errorf("  .Placeholder = %v; want %v", gotField.Placeholder, wantField.Placeholder)
				}
				if gotField.Value != wantField.Value {
					t.Errorf("  .Value = %v; want %v", gotField.Value, wantField.Value)
				}
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
