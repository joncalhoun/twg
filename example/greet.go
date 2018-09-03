// Package example is a package for demonstrating examples in source code.
package example

import "fmt"

type Demo struct{}

func (d Demo) Hello() {}

// Hello prints out hello to the person provided
func Hello(name string) (string, error) {
	return fmt.Sprintf("Hello, %s", name), nil
}

// Page will print out a message asking each person who hasn't checked in
// to do so.
func Page(checkIns map[string]bool) {
	for name, checkIn := range checkIns {
		if !checkIn {
			fmt.Printf("Paging %s; please see the front desk to check in.\n", name)
		}
	}
}
