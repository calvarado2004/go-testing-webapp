package main

import "net/url"

// errors is a map of field names to a slice of error messages.
type errors map[string][]string

// Get returns the first error message for the given field from the errors map.
func (e errors) Get(field string) string {

	errorSlice := e[field]

	if len(errorSlice) == 0 {
		return ""
	}

	return errorSlice[0]

}

// Add adds an error message for a given field to the errors map.
func (e errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

// Form represents an HTML form.
type Form struct {
	Data   url.Values
	Errors errors
}

// NewForm creates a new Form struct containing the provided form data.
func NewForm(data url.Values) *Form {
	return &Form{
		Data:   data,
		Errors: map[string][]string{},
	}
}

// Has checks if the form data contains the provided field.
func (f *Form) Has(field string) bool {
	x := f.Data.Get(field)
	if x == "" {
		return false
	}
	return true
}

// Required checks if the provided fields are present in the form data
func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Data.Get(field)
		if value == "" {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}

// MinLength checks if the provided field is at least a specific length.
func (f *Form) MinLength(field string, d int) bool {
	value := f.Data.Get(field)
	if value != "" && len(value) < d {
		f.Errors.Add(field, "This field is too short")
		return false
	}

	return true
}

// Check checks if the provided condition is true. If it is not, it adds an error
func (f *Form) Check(ok bool, key, message string) {
	if !ok {
		f.Errors.Add(key, message)
	}
}

// Valid returns true if there are no errors.
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

// IsEmail checks if the provided field is a valid email address.
func (f *Form) IsEmail(field string) {
	value := f.Data.Get(field)
	if value != "" {
		if len(value) < 3 || len(value) > 254 {
			f.Errors.Add(field, "This field must be a valid email address")
			return
		}
		if value[0] == '.' || value[len(value)-1] == '.' {
			f.Errors.Add(field, "This field must be a valid email address")
			return
		}
		for i := 0; i < len(value); i++ {
			switch value[i] {
			case '@':
				// OK as special case
			case '.':
				// OK if not first or last char
				if i == 0 || i == len(value)-1 {
					f.Errors.Add(field, "This field must be a valid email address")
					return
				}
			default:
				if value[i] < '0' || value[i] > 'z' {
					f.Errors.Add(field, "This field must be a valid email address")
					return
				}
			}
		}
	}
}
