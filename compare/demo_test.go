package compare

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

// See <https://golang.org/ref/spec#Comparison_operators> for more on comparisons

func TestSquare(t *testing.T) {
	arg := 4
	want := 16
	got := Square(arg)
	if got != want {
		t.Errorf("Square(%d) = %d; want %d", arg, got, want)
	}
}

func TestDog(t *testing.T) {
	morty := Dog{
		Name: "Morty",
		Age:  3,
	}
	morty2 := Dog{
		Name: "Morty",
		Age:  3,
	}
	t.Logf("morty=%p, morty2=%p", &morty, &morty2)
	if morty != morty2 {
		t.Errorf("morty != morty2")
	}
	odie := Dog{
		Name: "Odie",
		Age:  12,
	}
	if morty == odie {
		t.Errorf("morty == odie")
	}
}

func TestPtr(t *testing.T) {
	morty := &Dog{
		Name: "Morty",
		Age:  3,
	}
	morty2 := &Dog{
		Name: "Morty",
		Age:  3,
	}
	t.Logf("morty=%p, morty2=%p", morty, morty2)
	if morty == morty2 {
		t.Errorf("morty == morty2")
	}
}

func TestDogWithFn(t *testing.T) {
	morty := &DogWithFn{
		Name: "Morty",
		Age:  3,
	}
	morty2 := &DogWithFn{
		Name: "Morty",
		Age:  3,
	}
	if !reflect.DeepEqual(morty, morty2) {
		t.Errorf("morty != morty2")
	}
}

func TestAddTechniques(t *testing.T) {
	morty := &DogWithFn{
		Name: "Morty",
		Age:  1,
	}
	morty2 := &DogWithFn{
		Name: "Morty",
		Age:  3,
	}

	// Third party packages - see go-cmp

	// Check specific attributes you care about
	if morty.Name != morty2.Name {

	}

	// Helper function example: https://golang.org/src/net/http/httptest/recorder_test.go
	type checkFn func(*http.Response) error
	hasStatusRange := func(lo, hi int) checkFn {
		return func(res *http.Response) error {
			if res.StatusCode < lo || res.StatusCode > hi {
				return fmt.Errorf("status code = %d; wanted %d <= code <= %d", res.StatusCode, lo, hi)
				// status code = 333; wanted 200 <= code <= 299
			}
		}
	}

	tests := []struct {
		args     string
		checkFns []checkFn
	}{
		// test cases
		{"args", []checkFn{
			hasName("morty"),
			hasStatusCode(200),
		}},
		{"args", []checkFn{
			hasName("morty"),
			hasStatusRange(200, 299),
		}},
	}
	for _, tc := range tests {
		t.Run("...", func(t *testing.T) {
			got := SomeFunc(args)
			for _, check := range tc.checkFns {
				err := check(got)
				if err != nil {
					t.Error(err)
				}
			}
		})
	}

}
