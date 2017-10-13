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
	"github.com/theplant/validator/proto"
)

type Validate struct {
	GPValidate           *validator.Validate
	customTemplateMap    TemplateMap
	inclusionValidations map[string][]interface{}
}

type Rule struct {
	// Field mean field name of struct, it can be nested.
	// For example "Address.City".
	Field string
	// This tag contains tag and param, use "," to separate multiple tags.
	// For example "required,lte=20".
	Tag     string
	Code    string
	Message string
	Err     error
}

type Error struct {
	Field   string
	Tag     string
	Param   string
	Code    string
	Message string
	Err     error
}

type Errors []Error

type MapError map[string][]string

type TemplateMap map[string]string

var defaultTemplateMap = TemplateMap{
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

	inclusionValidations := map[string][]interface{}{}
	if err := gpValidate.RegisterValidation("inclusion", validateInclusion(inclusionValidations)); err != nil {
		panic(errors.Wrap(err, "register validation inclusion failed"))
	}

	validate := Validate{GPValidate: gpValidate, inclusionValidations: inclusionValidations}

	if err := validate.RegisterRegexpValidation("zipcode_jp", `^\d{3}-\d{4}$`); err != nil {
		panic(errors.Wrap(err, "register regexp validation zipcode_jp failed"))
	}

	if err := validate.RegisterRegexpValidation("simple_email", `^[^\s@]+@[^\s@]+$`); err != nil {
		panic(errors.Wrap(err, "register regexp validation simple_email failed"))
	}

	if err := validate.RegisterValidation("strict_required", validateStrictRequired); err != nil {
		panic(errors.Wrap(err, "register validation strict_required failed"))
	}

	return &validate
}

func validateStrictRequired(fl validator.FieldLevel) bool {
	return strings.TrimSpace(fl.Field().String()) != ""
}

// RegisterInclusionValidationParam register a param for inclusion validation.
// validSlice are all valid values for param of the inclusion validation,
// and it must be slice type.
//
// For example, if you register a "gender" param,
// then you can use `validate:"inclusion=gender"` validation tag for the struct.
//
// If you use a unregistered inclusion param,
// then this field of the struct validation always failed.
//
// If you register the same param multiple times, the front will be covered.
// If param is empty, it will return error.
func (v *Validate) RegisterInclusionValidationParam(param string, validSlice interface{}) error {
	if param == "" {
		return errors.New("param can not be empty")
	}

	validSliceVal := reflect.ValueOf(validSlice)

	if validSliceVal.Kind() != reflect.Slice {
		return errors.New("validSlice must be slice type")
	}

	validSliceIfaces := []interface{}{}
	for i := 0; i < validSliceVal.Len(); i++ {
		validSliceIfaces = append(validSliceIfaces, validSliceVal.Index(i).Interface())
	}

	v.inclusionValidations[param] = validSliceIfaces

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

func (ves Errors) Error() string {
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

// fmt []string{"1", "2", "3"} to `["1", "2", "3"]`
func fmtStringArray(strs []string) string {
	if len(strs) == 0 {
		return ""
	}

	return `["` + strings.Join(strs, `", "`) + `"]`
}

func (vem MapError) Error() string {
	errStr := ""

	for field, messages := range vem {
		errStr = errStr + field + ":" + fmtStringArray(messages) + " "
	}

	if errStr != "" {
		// Remove last " " char.
		errStr = errStr[:len(errStr)-1]
	}

	return errStr
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
func (v *Validate) DoRules(data interface{}, rules []Rule) (Errors, error) {
	return v.DoRulesWithTagName(data, rules, "")
}

func (v *Validate) DoRulesWithTagName(data interface{}, rules []Rule, tagName string) (verrs Errors, err error) {
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

	verrs = Errors{}

	for _, rule := range rules {
		field, fieldName := fieldByNameNested(val, rule.Field, tagName)
		if (field.Kind() == reflect.Invalid) || (field.Kind() == reflect.Ptr && field.IsNil()) {
			return nil, errors.New(fmt.Sprintf("get value from %v field failed", rule.Field))
		}
		fieldVal := field.Interface()

		varTags := []string{}
		for _, tag := range splitTag(rule.Tag) {
			switch getTagBefore(tag) {
			case "eqfield":
				otherField, _ := fieldByNameNested(val, getTagAfter(tag), "")
				if (otherField.Kind() == reflect.Invalid) || (otherField.Kind() == reflect.Ptr && otherField.IsNil()) {
					return nil, errors.New(fmt.Sprintf("get value from %v field failed", rule.Field))
				}
				otherFieldVal := otherField.Interface()

				verrs, err = appendErrors(v.GPValidate.VarWithValue(fieldVal, otherFieldVal, "eqfield"), verrs, fieldName, rule.Code, rule.Message, rule.Err)
				if err != nil {
					return nil, err
				}
			default:
				varTags = append(varTags, tag)
			}
		}

		if len(varTags) > 0 {
			verrs, err = appendErrors(v.GPValidate.Var(fieldVal, strings.Join(varTags, tagSeparator)), verrs, fieldName, rule.Code, rule.Message, rule.Err)
			if err != nil {
				return nil, err
			}
		}
	}

	if len(verrs) == 0 {
		return nil, nil
	}

	return verrs, nil
}

func setVErrsToStruct(verrs Errors, toStruct interface{}) {
	messageMap := map[string][]string{}
	errMap := map[string][]error{}
	for _, verr := range verrs {
		messageMap[verr.Field] = append(messageMap[verr.Field], verr.Message)
		errMap[verr.Field] = append(errMap[verr.Field], verr.Err)
	}

	v := reflect.ValueOf(toStruct).Elem()
	if v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}
	vv := v.Elem()

	for field, messages := range messageMap {
		vvf := vv.FieldByName(field)
		messagesV := reflect.ValueOf(messages)
		if vvf.Type().String() == messagesV.Type().String() {
			vvf.Set(reflect.ValueOf(messages))
		}
	}
	for field, errs := range errMap {
		vvf := vv.FieldByName(field)
		errsV := reflect.ValueOf(errs)
		if vvf.Type().String() == errsV.Type().String() {
			vvf.Set(reflect.ValueOf(errs))
		}
	}
}

func setToStructToNil(toStruct interface{}) {
	v := reflect.ValueOf(toStruct)
	v.Elem().Set(reflect.Zero(v.Elem().Type()))
}

func checkToStruct(toStruct interface{}) {
	v := reflect.ValueOf(toStruct)
	if v.Kind() != reflect.Ptr {
		panic(errors.New("toStruct must be pointer to pointer to struct"))
	}
	vv := v.Elem()
	if vv.Kind() != reflect.Ptr {
		panic(errors.New("toStruct must be pointer to pointer to struct"))
	}
}

// toStruct must be pointer to pointer to struct,
// and fields must contains all rule.Field and type must be []string.
// If no validation errors, then set toStruct to nil.
func (v *Validate) DoRulesToStructAndSetNil(data interface{}, rules []Rule, toStruct interface{}) {
	checkToStruct(toStruct)

	verrs, err := v.DoRules(data, rules)
	if err != nil {
		panic(err)
	}

	if len(verrs) == 0 {
		setToStructToNil(toStruct)
		return
	}

	setVErrsToStruct(verrs, toStruct)
}

// TODO toStruct can be pointer to struct.
// toStruct must be pointer to pointer to struct,
// and fields must contains all rule.Field and type must be []string.
func (v *Validate) DoRulesToStruct(data interface{}, rules []Rule, toStruct interface{}) {
	checkToStruct(toStruct)

	verrs, err := v.DoRules(data, rules)
	if err != nil {
		panic(err)
	}

	setVErrsToStruct(verrs, toStruct)
}

func verrsToProtoError(verrs Errors) (protoErr proto.Error) {
	for _, verr := range verrs {
		protoErr.FieldViolations = append(
			protoErr.FieldViolations,
			&proto.ValidationError_FieldViolation{
				Field:   verr.Field,
				Code:    verr.Code,
				Param:   verr.Param,
				Message: verr.Message,
			},
		)
	}

	return protoErr
}

// If no any error, BadRequest.FieldViolations == nil.
func (v *Validate) DoRulesToProtoError(data interface{}, rules []Rule) proto.Error {
	verrs, err := v.DoRules(data, rules)
	if err != nil {
		panic(err)
	}

	return verrsToProtoError(verrs)
}

func appendErrors(err error, verrs Errors, fieldName string, code string, message string, ruleErr error) (Errors, error) {
	if _, ok := err.(*validator.InvalidValidationError); ok {
		return nil, err
	}

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, validationErr := range validationErrors {
			verrs = append(verrs, Error{
				Field:   fieldName,
				Tag:     validationErr.Tag(),
				Param:   validationErr.Param(),
				Code:    code,
				Message: message,
				Err:     ruleErr,
			})
		}
	}

	return verrs, nil
}

// It's a proxy function for validate.Var from github.com/go-playground/validator.
// But it return bool type.
//
// If has some invalid input, it will return false.
func (v *Validate) IsVar(field interface{}, tag string) bool {
	err := v.GPValidate.Var(field, tag)

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

func getTemplate(tag string, customTemplateMap TemplateMap) string {
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
func VErrorsToMap(verrs Errors, templateMap TemplateMap) (MapError, error) {
	verrMap := MapError{}
	for _, verr := range verrs {
		vMessage, err := parseTemplate(templateValues{Param: verr.Param, Tag: verr.Tag}, getTemplate(verr.Tag, templateMap))
		if err != nil {
			return nil, errors.Wrap(err, "parseTemplate failed")
		}
		verrMap[verr.Field] = append(verrMap[verr.Field], vMessage)
	}
	return verrMap, nil
}

func (v *Validate) RegisterTemplateMap(templateMap TemplateMap) error {
	if err := checkTemplateMap(templateMap); err != nil {
		return err
	}

	v.customTemplateMap = templateMap

	return nil
}

func checkTemplateMap(templateMap TemplateMap) error {
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
func (v *Validate) DoRulesAndToMapError(data interface{}, rules []Rule) (MapError, error) {
	verrs, err := v.DoRules(data, rules)
	if err != nil {
		return nil, err
	}

	return VErrorsToMap(verrs, v.customTemplateMap)
}

func (v *Validate) DoRulesAndToMapErrorWithTagName(data interface{}, rules []Rule, tagName string) (MapError, error) {
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
	return v.GPValidate.RegisterValidation(tag, fn)
}

// RegisterRegexpValidation adds a regexp validation with the given tag and regexpString
//
// NOTES:
// - if the key already exists, the previous validation function will be replaced.
// - this method is not thread-safe it is intended that these all be registered prior to any validation
func (v *Validate) RegisterRegexpValidation(tag string, regexpString string) error {
	return v.GPValidate.RegisterValidation(tag, generateRegexpValidation(regexpString))
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

func validateInclusion(inclusionValidations map[string][]interface{}) validator.Func {
	return func(fl validator.FieldLevel) bool {
		return isInSlice(fl.Field().Interface(), inclusionValidations[fl.Param()])
	}
}

func isInSlice(check interface{}, array []interface{}) bool {
	for _, v := range array {
		if v == check {
			return true
		}
	}

	return false
}
