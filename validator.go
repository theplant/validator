package validator

import (
	"bytes"
	"fmt"
	"html/template"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator"
	"github.com/pkg/errors"
)

type Validate struct {
	gpValidate           *validator.Validate
	customTemplateMap    map[string]string
	inclusionValidations map[string][]string
}

type Rule struct {
	// Field mean field name of struct, it can be nested.
	// For example "Address.City".
	Field string
	// This tag contains tag and param, use "," to separate multiple tags.
	// For example "required,lte=20".
	Tag     string
	Message string
}

type VError struct {
	Field   string
	Tag     string
	Param   string
	Message string
}

type VErrors []VError

var defaultTemplateMap = map[string]string{
	"required":     "can not be blank",
	"lte":          "is too long, maximum length is {{.Param}}",
	"gte":          "is too short, minimum length is {{.Param}}",
	"max":          "is too large, maximum is {{.Param}}",
	"min":          "is too small, minimum is {{.Param}}",
	"zipcode_jp":   "invalid zipcode format, format is 123-1234",
	"inclusion":    "invalid {{.Param}} value",
	"simple_email": "invalid email format",
	"default":      "validation failed with {{ if eq .Param \"\" }}{{.Tag}}{{ else }}{{.Tag}}={{.Param}}{{ end }}",
}

func New() *Validate {
	gpValidate := validator.New()

	inclusionValidations := map[string][]string{}
	if err := gpValidate.RegisterValidation("inclusion", validateInclusion(inclusionValidations)); err != nil {
		panic(errors.Wrap(err, "register validation inclusion failed"))
	}

	validate := Validate{gpValidate: gpValidate, inclusionValidations: inclusionValidations}

	if err := validate.RegisterRegexpValidation("zipcode_jp", `^\d{3}-\d{4}$`); err != nil {
		panic(errors.Wrap(err, "register regexp validation zipcode_jp failed"))
	}

	if err := validate.RegisterRegexpValidation("simple_email", `^[^\s@]+@[^\s@]+$`); err != nil {
		panic(errors.Wrap(err, "register regexp validation simple_email failed"))
	}

	return &validate
}

// RegisterInclusionValidationParam register a param for inclusion validation.
// validList are all valid values for param of the inclusion validation.
//
// For example, if you register a "gender" param,
// then you can use `validate:"inclusion=gender"` validation tag for the struct.
//
// If you use a unregistered inclusion param,
// then this field of the struct validation always failed.
//
// If you register the same param multiple times, the front will be covered.
// If param is empty, it will return error.
func (v *Validate) RegisterInclusionValidationParam(param string, validList []string) error {
	if param == "" {
		return errors.New("param can not be empty")
	}

	v.inclusionValidations[param] = validList

	return nil
}

// val should be a struct value.
// If found the tagName of val, then use the tag value replace name.
//
// It return invalid value if has some errors.
// It return nil pointer value if try to get value from nil.
func fieldByNameNested(val reflect.Value, name string, tagName string) (reflect.Value, string) {
	names := []string{}

	for _, n := range strings.Split(name, ".") {
		if val.Kind() == reflect.Invalid {
			return val, ""
		}
		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				return val, ""
			}
			val = val.Elem()
		}

		tag := getTagValue(val, n, tagName)
		if tag == "" {
			names = append(names, n)
		} else {
			names = append(names, tag)
		}

		val = val.FieldByName(n)
	}

	return val, strings.Join(names, ".")
}

// If get failed, then return "".
// If tag value == "-", then return "".
func getTagValue(val reflect.Value, name, tagName string) string {
	field, ok := val.Type().FieldByName(name)
	if !ok {
		return ""
	}

	flatTag := field.Tag.Get(tagName)
	if flatTag == "-" {
		return ""
	}

	tag := ""
	tags := strings.Split(flatTag, ",")
	if len(tags) > 0 {
		tag = tags[0]
	} else {
		tag = flatTag
	}

	return tag
}

func (ves VErrors) Error() string {
	if len(ves) == 0 {
		return ""
	}

	errStrs := []string{}
	for _, ve := range ves {
		if ve.Param == "" {
			if ve.Message == "" {
				errStrs = append(errStrs, fmt.Sprintf("%v of %v", ve.Tag, ve.Field))
			} else {
				errStrs = append(errStrs, fmt.Sprintf("%v of %v: %v", ve.Tag, ve.Field, ve.Message))
			}
		} else {
			if ve.Message == "" {
				errStrs = append(errStrs, fmt.Sprintf("%v=%v of %v", ve.Tag, ve.Param, ve.Field))
			} else {
				errStrs = append(errStrs, fmt.Sprintf("%v=%v of %v: %v", ve.Tag, ve.Param, ve.Field, ve.Message))
			}
		}
	}

	return "validation failed: " + strings.Join(errStrs, "; ")
}

// data should be a struct or a pointer to struct.
//
// if return (nil, nil), it mean no validation error.
//
// If it return (nil, error), you must to solve it. Possible errors:
// * Invalid Rule.Tag
// * Invalid Rule.Field
// * data is not a struct or a pointer to struct
//
// Some custom tags:
// * zipcode_jp
// * simple_email
func (v *Validate) DoRules(data interface{}, rules []Rule) (verrs VErrors, err error) {
	return v.DoRulesWithTagName(data, rules, "")
}

func (v *Validate) DoRulesWithTagName(data interface{}, rules []Rule, tagName string) (verrs VErrors, err error) {
	defer func() {
		if r := recover(); r != nil {
			verrs = nil
			err = errors.New(fmt.Sprint(r))
		}
	}()

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr && !val.IsNil() {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, errors.New("data should be a struct or a pointer to struct")
	}

	verrs = VErrors{}

	for _, rule := range rules {
		field, fieldName := fieldByNameNested(val, rule.Field, tagName)
		fieldVal := field.Interface()
		if (field.Kind() == reflect.Invalid) || (val.Kind() == reflect.Ptr && val.IsNil()) {
			return nil, errors.New(fmt.Sprintf("get value from %v field failed", rule.Field))
		}

		err := v.gpValidate.Var(fieldVal, rule.Tag)

		if _, ok := err.(*validator.InvalidValidationError); ok {
			return nil, err
		}

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, validationErr := range validationErrors {
				verrs = append(verrs, VError{
					Field:   fieldName,
					Tag:     validationErr.Tag(),
					Param:   validationErr.Param(),
					Message: rule.Message,
				})
			}
		}
	}

	if len(verrs) == 0 {
		return nil, nil
	}

	return verrs, nil
}

// It's a proxy function for validate.Var from github.com/go-playground/validator.
// But it return bool type.
//
// If has some invalid input, it will return false.
func (v *Validate) IsVar(field interface{}, tag string) bool {
	err := v.gpValidate.Var(field, tag)

	if _, ok := err.(*validator.InvalidValidationError); ok {
		return false
	}

	if _, ok := err.(validator.ValidationErrors); ok {
		return false
	}

	return true
}

type templateValues struct {
	Param string
	Tag   string
}

func getTemplate(tag string, customTemplateMap map[string]string) string {
	if message := customTemplateMap[tag]; message != "" {
		return message
	}

	if message := defaultTemplateMap[tag]; message != "" {
		return message
	}

	return defaultTemplateMap["default"]
}

func parseTemplate(tplValues templateValues, templateStr string) (string, error) {
	tl, err := template.New("validate").Parse(templateStr)
	if err != nil {
		return "", errors.Wrap(err, "template parse failed")
	}

	message := bytes.Buffer{}
	if err := tl.Execute(&message, tplValues); err != nil {
		return "", errors.Wrap(err, "template execute failed")
	}

	return message.String(), nil
}

// templateMap is a map that mean [tag]template,
// templateMap will be parsed by go template,
// you can use ".Tag" and ".Param" variable in the template.
//
// For example, validation tag is "max=100", then tag is "max", param is "100",
// if template is "is too large, maximum is {{.Param}}",
// then it will be parsed to "is too large, maximum is 100".
//
// If templateMap is nil, it will use defaultTemplateMap to parse.
// If not found the tag in the templateMap, it will use defaultTemplateMap to parse for this tag,
// it mean templateMap will merge to defaultTemplateMap,
// so you don't worry about missing some tags of the templateMap.
//
// If parse template failed, it will return error.
func VErrorsToMap(verrs VErrors, templateMap map[string]string) (map[string][]string, error) {
	vMap := map[string][]string{}
	for _, verr := range verrs {
		vMessage, err := parseTemplate(templateValues{Param: verr.Param, Tag: verr.Tag}, getTemplate(verr.Tag, templateMap))
		if err != nil {
			return nil, errors.Wrap(err, "parseTemplate failed")
		}
		vMap[verr.Field] = append(vMap[verr.Field], vMessage)
	}
	return vMap, nil
}

func (v *Validate) RegisterTemplateMap(templateMap map[string]string) error {
	if err := checkTemplateMap(templateMap); err != nil {
		return err
	}

	v.customTemplateMap = templateMap

	return nil
}

func checkTemplateMap(templateMap map[string]string) error {
	tplValues := templateValues{
		Param: "check param",
		Tag:   "check tag",
	}

	for tag, tpl := range templateMap {
		if tag == "" {
			return errors.New("tag of the templateMap can not be empty")
		}

		_, err := parseTemplate(tplValues, tpl)
		if err != nil {
			return errors.Wrap(err, "tpl of the templateMap invalid")
		}
	}

	return nil
}

// Same as DoRules, run DoRules and VErrorsToMap with custom template.
// You can use RegisterTemplateMap func to register custom template.
func (v *Validate) DoRulesAndToMap(data interface{}, rules []Rule) (map[string][]string, error) {
	verrs, err := v.DoRules(data, rules)
	if err != nil {
		return nil, err
	}

	return VErrorsToMap(verrs, v.customTemplateMap)
}

func (v *Validate) DoRulesAndToMapWithTagName(data interface{}, rules []Rule, tagName string) (map[string][]string, error) {
	verrs, err := v.DoRulesWithTagName(data, rules, tagName)
	if err != nil {
		return nil, err
	}

	return VErrorsToMap(verrs, v.customTemplateMap)
}

// RegisterValidation adds a validation with the given tag.
//
// NOTES:
// - if the key already exists, the previous validation function will be replaced.
// - this method is not thread-safe it is intended that these all be registered prior to any validation
//
// TODO if we use this function, we must import github.com/go-playground/validator,
// this is not good, because we hope only import this package.
// To find a better way.
func (v *Validate) RegisterValidation(tag string, fn func(validator.FieldLevel) bool) error {
	return v.gpValidate.RegisterValidation(tag, fn)
}

// RegisterRegexpValidation adds a regexp validation with the given tag and regexpString
//
// NOTES:
// - if the key already exists, the previous validation function will be replaced.
// - this method is not thread-safe it is intended that these all be registered prior to any validation
func (v *Validate) RegisterRegexpValidation(tag string, regexpString string) error {
	return v.gpValidate.RegisterValidation(tag, generateRegexpValidation(regexpString))
}

func generateRegexpValidation(regexpString string) func(fl validator.FieldLevel) bool {
	return func(fl validator.FieldLevel) bool {
		val := fl.Field().String()

		if !regexp.MustCompile(regexpString).MatchString(val) {
			return false
		}

		return true
	}
}

func validateInclusion(inclusionValidations map[string][]string) validator.Func {
	return func(fl validator.FieldLevel) bool {
		return isInStringArray(fl.Field().String(), inclusionValidations[fl.Param()])
	}
}

func isInStringArray(check string, array []string) bool {
	for _, v := range array {
		if v == check {
			return true
		}
	}

	return false
}
