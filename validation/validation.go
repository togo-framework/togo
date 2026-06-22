// Package validation is a small, Laravel-style request validator driven by
// `validate` struct tags. Generated REST handlers run it on the request body so
// API inputs are checked before they reach the database.
//
//	type CreatePost struct {
//	    Title string `json:"title" validate:"required,min=3,max=120"`
//	    Email string `json:"email" validate:"required,email"`
//	}
//	if errs := validation.Validate(in.Body); errs.Any() { ... 422 ... }
package validation

import (
	"fmt"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Errors maps a field name to its failed-rule messages.
type Errors map[string][]string

// Any reports whether any field failed.
func (e Errors) Any() bool { return len(e) > 0 }

// Error renders all failures as a single message (implements error).
func (e Errors) Error() string {
	parts := make([]string, 0, len(e))
	for field, msgs := range e {
		parts = append(parts, field+": "+strings.Join(msgs, ", "))
	}
	return strings.Join(parts, "; ")
}

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-?[0-9a-fA-F]{4}-?[0-9a-fA-F]{4}-?[0-9a-fA-F]{4}-?[0-9a-fA-F]{12}$`)

// Validate checks v (a struct or pointer) against its `validate` tags.
func Validate(v any) Errors {
	errs := Errors{}
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return errs
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return errs
	}
	t := rv.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("validate")
		if tag == "" || tag == "-" {
			continue
		}
		name := jsonName(f)
		fv := rv.Field(i)
		for _, rule := range strings.Split(tag, ",") {
			if msg := check(strings.TrimSpace(rule), fv); msg != "" {
				errs[name] = append(errs[name], msg)
			}
		}
	}
	return errs
}

func check(rule string, fv reflect.Value) string {
	name, arg, _ := strings.Cut(rule, "=")
	// Dereference pointers (nullable fields).
	for fv.Kind() == reflect.Ptr {
		if fv.IsNil() {
			if name == "required" {
				return "is required"
			}
			return "" // other rules skip absent optional values
		}
		fv = fv.Elem()
	}
	switch name {
	case "required":
		if isZero(fv) {
			return "is required"
		}
	case "email":
		if s := str(fv); s != "" {
			if _, err := mail.ParseAddress(s); err != nil {
				return "must be a valid email"
			}
		}
	case "url":
		if s := str(fv); s != "" {
			if u, err := url.ParseRequestURI(s); err != nil || u.Scheme == "" {
				return "must be a valid URL"
			}
		}
	case "uuid":
		if s := str(fv); s != "" && !uuidRe.MatchString(s) {
			return "must be a valid UUID"
		}
	case "min":
		if n, _ := strconv.Atoi(arg); measure(fv) < n {
			return fmt.Sprintf("must be at least %s", arg)
		}
	case "max":
		if n, _ := strconv.Atoi(arg); measure(fv) > n {
			return fmt.Sprintf("must be at most %s", arg)
		}
	case "len":
		if n, _ := strconv.Atoi(arg); measure(fv) != n {
			return fmt.Sprintf("must be exactly %s", arg)
		}
	case "in":
		opts := strings.Split(arg, " ")
		s := str(fv)
		for _, o := range opts {
			if o == s {
				return ""
			}
		}
		return "must be one of: " + strings.ReplaceAll(arg, " ", ", ")
	case "numeric":
		if s := str(fv); s != "" {
			if _, err := strconv.ParseFloat(s, 64); err != nil {
				return "must be numeric"
			}
		}
	}
	return ""
}

// measure returns length for strings/slices/maps, else the numeric value.
func measure(fv reflect.Value) int {
	switch fv.Kind() {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		return fv.Len()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int(fv.Int())
	case reflect.Float32, reflect.Float64:
		return int(fv.Float())
	}
	return 0
}

func isZero(fv reflect.Value) bool { return fv.IsZero() }

func str(fv reflect.Value) string {
	if fv.Kind() == reflect.String {
		return fv.String()
	}
	return ""
}

func jsonName(f reflect.StructField) string {
	if tag := f.Tag.Get("json"); tag != "" && tag != "-" {
		if name, _, _ := strings.Cut(tag, ","); name != "" {
			return name
		}
	}
	return f.Name
}
