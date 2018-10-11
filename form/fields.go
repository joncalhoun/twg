package form

import (
	"reflect"
	"strings"
)

func valueOf(v interface{}) reflect.Value {
	var rv reflect.Value
	switch value := v.(type) {
	case reflect.Value:
		rv = value
	default:
		rv = reflect.ValueOf(v)
	}

	// With any pointers we really want to just work with their underlying
	// type.
	if rv.Kind() == reflect.Ptr {
		// The underlying type is pretty useless if it is nil, so we need to
		// instantiate a new copy of whatever that is before using it.
		if rv.IsNil() {
			rv = reflect.New(rv.Type().Elem())
		}
		rv = rv.Elem()
	}
	return rv
}

func fields(strct interface{}, parentNames ...string) []field {
	rv := valueOf(strct)
	if rv.Kind() != reflect.Struct {
		panic("form: invalid value; only structs are supported")
	}
	t := rv.Type()
	var ret []field
	for i := 0; i < t.NumField(); i++ {
		tf := t.Field(i)
		rvf := valueOf(rv.Field(i))
		if !rvf.CanInterface() {
			continue
		}
		if rvf.Kind() == reflect.Struct {
			nestedParentNames := append(parentNames, tf.Name)
			nestedFields := fields(rvf.Interface(), nestedParentNames...)
			ret = append(ret, nestedFields...)
			continue
		}
		names := append(parentNames, tf.Name)
		name := strings.Join(names, ".")
		f := field{
			Label:       tf.Name,
			Name:        name,
			Type:        "text",
			Placeholder: tf.Name,
			Value:       rvf.Interface(),
		}
		f.apply(parseTags(tf))
		ret = append(ret, f)
	}
	return ret
}

type field struct {
	Label       string
	Name        string
	Type        string
	Placeholder string
	Value       interface{}
	Errors      []string
}

func (f *field) apply(tags map[string]string) {
	if v, ok := tags["name"]; ok {
		f.Name = v
	}
	if v, ok := tags["label"]; ok {
		f.Label = v
	}
	if v, ok := tags["placeholder"]; ok {
		f.Placeholder = v
	}
	if v, ok := tags["type"]; ok {
		f.Type = v
	}
}

func (f *field) setErrors(errors []FieldError) {
	for _, fe := range errors {
		if fe.Field == f.Name {
			f.Errors = append(f.Errors, fe.Error)
		}
	}
}

func parseTags(sf reflect.StructField) map[string]string {
	// label=Full Name;name=full_name
	rawTag := sf.Tag.Get("form")
	if len(rawTag) == 0 {
		return nil
	}
	ret := make(map[string]string)
	// tags = [label=Full Name, name=full_name]
	tags := strings.Split(rawTag, ";")
	for _, tag := range tags {
		kv := strings.Split(tag, "=")
		if len(kv) != 2 {
			panic("form: invalid struct tag")
		}
		k, v := kv[0], kv[1]
		ret[k] = v
	}
	return ret
}
