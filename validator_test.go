package validator_test

import (
	"reflect"
	"regexp"
	"testing"

	gpvalidator "github.com/go-playground/validator"
	"github.com/pkg/errors"
	"github.com/theplant/aigle/validator"
)

type info struct {
	Name     string
	Password string
	Age      int
	Address  string
	ZipCode  string
}

var infoRules = []validator.Rule{
	{Field: "Name", Tag: "required,lte=20"},
	{Field: "Password", Tag: "gte=8"},
	{Field: "Age", Tag: "min=20,max=100"},
	{Field: "Address", Tag: "required,lte=50"},
	{Field: "ZipCode", Tag: "zipcode_jp"},
}

func TestValidate_DoRulesWithNestedStruct(t *testing.T) {
	type address struct {
		City    string
		Address string
	}

	type user struct {
		Name    string
		Age     int
		Address address
	}

	userRules := []validator.Rule{
		{Field: "Address.City", Tag: "required"},
	}

	validate, err := validator.New()
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	verrs, err := validate.DoRules(user{}, userRules)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	gotValidationErrs, err := validator.VErrorsToMap(verrs, nil)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
	wantValidationErrs := map[string][]string{
		"Address.City": {"can not be blank"},
	}

	if !reflect.DeepEqual(wantValidationErrs, gotValidationErrs) {
		t.Fatalf("got %v, want %v", gotValidationErrs, wantValidationErrs)
	}
}

func TestValidate_DoRulesWithRuleIsNil(t *testing.T) {
	validate, err := validator.New()
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	verrs, err := validate.DoRules(info{}, nil)

	if err != nil {
		t.Fatalf("should not return error: %v", err)
	}

	if verrs != nil {
		t.Fatalf("should not return VErrors: %v", verrs)
	}
}

func TestValidate_DoRulesWithInvalidTagAndParamOfRule(t *testing.T) {
	newInfoRules := []validator.Rule{
		{Field: "Name", Tag: "unknow tag"},
	}

	validate, err := validator.New()
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	_, err = validate.DoRules(info{}, newInfoRules)

	if err == nil {
		t.Fatal("should return error")
	}
}

func TestValidate_DoRulesWithInvalidFieldOfRule(t *testing.T) {
	newInfoRules := []validator.Rule{
		{Field: "filed does not exist", Tag: "required"},
	}

	validate, err := validator.New()
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	_, err = validate.DoRules(info{}, newInfoRules)

	if err == nil {
		t.Fatal("should return error")
	}
}

func TestValidate_DoRulesWithDataTypeError(t *testing.T) {
	validate, err := validator.New()
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	_, err = validate.DoRules("error type", infoRules)

	if err == nil {
		t.Fatal("should return error")
	}
}

func TestValidate_DoRulesWithMessageCanBeReturn(t *testing.T) {
	newInfoRules := []validator.Rule{
		{Field: "Name", Tag: "required", Message: "I'm a message"},
	}

	validate, err := validator.New()
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	gotVerrs, err := validate.DoRules(info{}, newInfoRules)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	if len(gotVerrs) != 1 {
		t.Fatal("should return one VError")
	}

	wantVerrs := validator.VError{
		Field:   "Name",
		Tag:     "required",
		Message: "I'm a message",
	}

	if !reflect.DeepEqual(gotVerrs[0], wantVerrs) {
		t.Fatalf("got %v, want %v", gotVerrs, wantVerrs)
	}
}

func TestValidate_RegisterInclusionValidationParam(t *testing.T) {
	type info struct {
		Gender  string
		Color   string
		Species string
	}

	infoRules := []validator.Rule{
		{Field: "Gender", Tag: "inclusion=gender"},
		{Field: "Color", Tag: "inclusion=color"},
		{Field: "Species", Tag: "inclusion=species"},
	}

	infoFailed := info{
		Gender:  "U",
		Color:   "black",
		Species: "panda",
	}

	validate, err := validator.New()
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	if err := validate.RegisterInclusionValidationParam("gender", []string{"U", "M", "F"}); err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
	if err := validate.RegisterInclusionValidationParam("color", []string{"blue", "red", "black"}); err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
	if err := validate.RegisterInclusionValidationParam("species", []string{"dog", "cat"}); err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	verrs, err := validate.DoRules(infoFailed, infoRules)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	gotValidationErrs, err := validator.VErrorsToMap(verrs, nil)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
	wantValidationErrs := map[string][]string{
		"Species": {"invalid species value"},
	}

	if !reflect.DeepEqual(wantValidationErrs, gotValidationErrs) {
		t.Fatalf("got %v, want %v", gotValidationErrs, wantValidationErrs)
	}
}

func TestValidate_RegisterInclusionValidationParamWithNotRegister(t *testing.T) {
	type info struct {
		Gender string
	}

	infoRules := []validator.Rule{
		{Field: "Gender", Tag: "inclusion=gender"},
	}

	infoFailed := info{
		Gender: "U",
	}

	validate, err := validator.New()
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	verrs, err := validate.DoRules(infoFailed, infoRules)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	gotValidationErrs, err := validator.VErrorsToMap(verrs, nil)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
	wantValidationErrs := map[string][]string{
		"Gender": {"invalid gender value"},
	}

	if !reflect.DeepEqual(wantValidationErrs, gotValidationErrs) {
		t.Fatalf("got %v, want %v", gotValidationErrs, wantValidationErrs)
	}
}

func TestValidate_RegisterInclusionValidationParam_ParamIsEmpty(t *testing.T) {
	validate, err := validator.New()
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	err = validate.RegisterInclusionValidationParam("", []string{"U", "M", "F"})
	if errors.Cause(err).Error() != "param can not be empty" {
		t.Fatal("should return `param can not be empty`")
	}
}

func TestVErrorsToMapWithDefaultTemplateMap(t *testing.T) {
	infoEmpty := info{}

	validate, err := validator.New()
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	verrs, err := validate.DoRules(infoEmpty, infoRules)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	gotValidationErrs, err := validator.VErrorsToMap(verrs, nil)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
	wantValidationErrs := map[string][]string{
		"Name":     {"can not be blank"},
		"Password": {"is too short, minimum length is 8"},
		"Age":      {"is too small, minimum is 20"},
		"Address":  {"can not be blank"},
		"ZipCode":  {"invalid zipcode format, format is 123-1234"},
	}

	if !reflect.DeepEqual(wantValidationErrs, gotValidationErrs) {
		t.Fatalf("got %v, want %v", gotValidationErrs, wantValidationErrs)
	}
}

func TestVErrorsToMapWithCustomTemplateMap(t *testing.T) {
	infoFailed := info{
		Name:     "1234567890 1234567890 1234567890",
		Password: "1234",
		Age:      200,
		Address:  "1234567890 1234567890 1234567890 1234567890 1234567890",
		ZipCode:  "1234-123",
	}
	templateMap := map[string]string{
		"lte":        "maximum length is {{.Param}}",
		"gte":        "minimum length is {{.Param}}",
		"zipcode_jp": "invalid zipcode format",
	}

	validate, err := validator.New()
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	verrs, err := validate.DoRules(infoFailed, infoRules)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	gotValidationErrs, err := validator.VErrorsToMap(verrs, templateMap)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
	wantValidationErrs := map[string][]string{
		"Name":     {"maximum length is 20"},
		"Password": {"minimum length is 8"},
		"Age":      {"is too large, maximum is 100"},
		"Address":  {"maximum length is 50"},
		"ZipCode":  {"invalid zipcode format"},
	}

	if !reflect.DeepEqual(wantValidationErrs, gotValidationErrs) {
		t.Fatalf("got %v, want %v", gotValidationErrs, wantValidationErrs)
	}
}

func TestVErrorsToMapWithDefaultFieldOfDefaultTemplate(t *testing.T) {
	type info struct {
		Name string
		IPv4 string
	}

	infoRules := []validator.Rule{
		{Field: "Name", Tag: "lt=20"},
		{Field: "IPv4", Tag: "ipv4"},
	}

	infoFailed := info{
		Name: "12345678901234567890",
		IPv4: "1271",
	}

	validate, err := validator.New()
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	verrs, err := validate.DoRules(infoFailed, infoRules)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	gotValidationErrs, err := validator.VErrorsToMap(verrs, nil)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
	wantValidationErrs := map[string][]string{
		"Name": {"validation failed with lt=20"},
		"IPv4": {"validation failed with ipv4"},
	}

	if !reflect.DeepEqual(wantValidationErrs, gotValidationErrs) {
		t.Fatalf("got %v, want %v", gotValidationErrs, wantValidationErrs)
	}
}

func TestVErrorsToMapWithParseTemplateError(t *testing.T) {
	verrs := validator.VErrors{
		{Field: "Name", Tag: "required"},
	}

	templateMap := map[string]string{
		"required": "maximum length is {{.Param",
	}
	_, err := validator.VErrorsToMap(verrs, templateMap)
	if err == nil {
		t.Fatal("should return err")
	}

	templateMap = map[string]string{
		"required": "maximum length is {{.otherVariable}}",
	}
	_, err = validator.VErrorsToMap(verrs, templateMap)
	if err == nil {
		t.Fatal("should return err")
	}
}

func TestValidate_RegisterValidation(t *testing.T) {
	type info struct {
		Phone string
	}

	infoRules := []validator.Rule{
		{Field: "Phone", Tag: "phone"},
	}

	infoFailed := info{
		Phone: "123",
	}

	validate, err := validator.New()
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	validatePhone := func(fl gpvalidator.FieldLevel) bool {
		phone := fl.Field().String()

		if !regexp.MustCompile(`^\d{3}-\d{4}-\d{4}$`).MatchString(phone) {
			return false
		}

		return true
	}
	if err := validate.RegisterValidation("phone", validatePhone); err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	verrs, err := validate.DoRules(infoFailed, infoRules)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	gotValidationErrs, err := validator.VErrorsToMap(verrs, nil)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
	wantValidationErrs := map[string][]string{
		"Phone": {"validation failed with phone"},
	}

	if !reflect.DeepEqual(wantValidationErrs, gotValidationErrs) {
		t.Fatalf("got %v, but want %v", gotValidationErrs, wantValidationErrs)
	}
}

func TestValidate_RegisterRegexpValidation(t *testing.T) {
	type info struct {
		Phone string
	}

	infoRules := []validator.Rule{
		{Field: "Phone", Tag: "phone"},
	}

	infoFailed := info{
		Phone: "----",
	}

	validate, err := validator.New()
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	if err := validate.RegisterRegexpValidation("phone", `^\d{0,5}-\d{0,5}-\d{0,5}$`); err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	verrs, err := validate.DoRules(infoFailed, infoRules)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	gotValidationErrs, err := validator.VErrorsToMap(verrs, nil)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
	wantValidationErrs := map[string][]string{
		"Phone": {"validation failed with phone"},
	}

	if !reflect.DeepEqual(wantValidationErrs, gotValidationErrs) {
		t.Fatalf("got %v, but want %v", gotValidationErrs, wantValidationErrs)
	}
}

func TestValidate_IsVar(t *testing.T) {
	validate, err := validator.New()
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	if !(validate.IsVar(-1, "lt=0") == true) {
		t.Fatal("should return true")
	}

	if !(validate.IsVar(0, "lt=0") == false) {
		t.Fatal("should return false")
	}

	if !(validate.IsVar(1, "lt=0") == false) {
		t.Fatal("should return false")
	}

	if !(validate.IsVar("127.0.0", "ipv4") == false) {
		t.Fatal("should return false")
	}

	if !(validate.IsVar("127.0.0.1", "ipv4") == true) {
		t.Fatal("should return true")
	}
}
