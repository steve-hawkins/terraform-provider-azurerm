package validate

import (
	"fmt"
	"regexp"
)

func SharedImageGalleryName(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)

	r, _ := regexp.Compile("^[A-Za-z0-9._-]+$")
	if !r.MatchString(value) {
		es = append(es, fmt.Errorf("%s can only contain alphanumeric, full stops, underscores and dashes. Got %q.", k, value))
	}

	length := len(value)
	if length >= 80 {
		es = append(es, fmt.Errorf("%s can be up to 80 characters, currently %d.", k, length))
	}

	return
}

func SharedImageName(v interface{}, k string) (ws []string, es []error) {
	// same as Shared Image Gallery name, for now.
	return SharedImageGalleryName(v, k)
}

func SharedImageVersionName(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)

	r, _ := regexp.Compile("^([0-9]\\.[0-9]\\.[0-9])$")
	if !r.MatchString(value) {
		es = append(es, fmt.Errorf("Expected %s to be in the format `1.2.3` but got %q.", k, value))
	}

	return
}
