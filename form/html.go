package form

import (
	"html/template"
	"strings"
)

// FieldError is provided as a way to denote errors with specific fields.
type FieldError struct {
	Field string
	Error string
}

// HTML is used to generate HTML forms/inputs from Go structs. Given a
// template that looks something like this:
//
//     <input type="{{.Type}}" name="{{.Name}}" {{with .Value}}value="{{.}}"{{end}}>
//
// And a struct like this:
//
//     struct{
//	     Name string
//       Email string
//     }{
// 			 Name: "Michael Scott",
//       Email: "michael@dundermifflin.com",
//     }
//
// The HTML function will return a template.HTML of:
//
//     <input type="text" name="Name" value="Michael Scott">
//     <input type="text" name="Email" value="michael@dundermifflin.com">
//
// An example similar to this is shown as the first test case in TestHTML
// in the html_test.go source file.
//
// Note: This does not currently support struct tags, but will eventually
// in order to support more customization and flexibility.
func HTML(t *template.Template, strct interface{}, errors ...FieldError) (template.HTML, error) {
	var inputs []string
	for _, field := range fields(strct) {
		field.setErrors(errors)
		var sb strings.Builder
		err := t.Execute(&sb, field)
		if err != nil {
			return "", err
		}
		inputs = append(inputs, sb.String())
	}
	return template.HTML(strings.Join(inputs, "")), nil
}
