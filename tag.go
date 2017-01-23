package validator

import (
	"reflect"
	"strings"
)

const (
	tagSeparator    = ","
	tagKeySeparator = "="
	tagIgnore       = "-"
)

// Return value length must be >=1.
func splitTag(tag string) []string {
	return strings.Split(tag, tagSeparator)
}

// Example, input lte=20, output lte
func getTagBefore(tag string) string {
	return strings.SplitN(tag, tagKeySeparator, 2)[0]
}

// Example, input lte=20, output 20
func getTagAfter(tag string) string {
	strs := strings.SplitN(tag, tagKeySeparator, 2)
	if len(strs) > 1 {
		return strs[1]
	}
	return strs[0]
}

// If get failed, then return "".
// If tag value == "-", then return "".
// If tag is "a,b,c", then return first value a.
func getTagValue(val reflect.Value, name, tagName string) string {
	field, ok := val.Type().FieldByName(name)
	if !ok {
		return ""
	}

	tag := field.Tag.Get(tagName)
	if tag == tagIgnore {
		return ""
	}

	return splitTag(tag)[0]
}
