package validator_test

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	gpvalidator "github.com/go-playground/validator"
	"github.com/pkg/errors"
	"github.com/theplant/testingutils"
	"github.com/theplant/validator"
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

type infoError struct {
	Name     []string
	Password []string
	Age      []string
	Address  []string
	ZipCode  []string
}

func TestValidate_DoRulesWithNested(t *testing.T) {
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

	validate := validator.New()

	verrs, err := validate.DoRules(user{}, userRules)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	gotValidationErrs, err := validator.VErrorsToMap(verrs, nil)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
	wantValidationErrs := validator.MapError{
		"Address.City": {"can not be blank"},
	}

	if !reflect.DeepEqual(wantValidationErrs, gotValidationErrs) {
		t.Fatalf("got %v, want %v", gotValidationErrs, wantValidationErrs)
	}
}

func TestValidate_DoRulesWithRuleIsNil(t *testing.T) {
	validate := validator.New()

	verrs, err := validate.DoRules(info{}, nil)

	if err != nil {
		t.Fatalf("should not return error: %v", err)
	}

	if verrs != nil {
		t.Fatalf("should not return Errors: %v", verrs)
	}
}

func TestValidate_DoRulesWithInvalidTagAndParamOfRule(t *testing.T) {
	newInfoRules := []validator.Rule{
		{Field: "Name", Tag: "unknow tag"},
	}

	validate := validator.New()

	_, err := validate.DoRules(info{}, newInfoRules)

	if err == nil {
		t.Fatal("should return error")
	}
}

func TestValidate_DoRulesWithInvalidFieldOfRule(t *testing.T) {
	newInfoRules := []validator.Rule{
		{Field: "filed does not exist", Tag: "required"},
	}

	validate := validator.New()

	_, err := validate.DoRules(info{}, newInfoRules)

	if err == nil {
		t.Fatal("should return error")
	}
}

func TestValidate_DoRulesWithDataTypeError(t *testing.T) {
	validate := validator.New()

	_, err := validate.DoRules("error type", infoRules)

	if err == nil {
		t.Fatal("should return error")
	}
}

func TestValidate_DoRulesWithMessageCanBeReturn(t *testing.T) {
	newInfoRules := []validator.Rule{
		{Field: "Name", Tag: "required", Message: "I'm a message"},
	}

	validate := validator.New()

	gotVerrs, err := validate.DoRules(info{}, newInfoRules)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	if len(gotVerrs) != 1 {
		t.Fatal("should return one Error")
	}

	wantVerrs := validator.Error{
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

	validate := validator.New()

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
	wantValidationErrs := validator.MapError{
		"Species": {"invalid species value"},
	}

	if !reflect.DeepEqual(wantValidationErrs, gotValidationErrs) {
		t.Fatalf("got %v, want %v", gotValidationErrs, wantValidationErrs)
	}
}

func TestValidate_RegisterInclusionValidationParamWithIntSliceType(t *testing.T) {
	type info struct {
		Number int
	}

	infoRules := []validator.Rule{
		{Field: "Number", Tag: "inclusion=number"},
	}

	validate := validator.New()

	if err := validate.RegisterInclusionValidationParam("number", []int{1, 2, 3}); err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	// validate failed

	infoFailed := info{
		Number: 100,
	}

	verrs, err := validate.DoRules(infoFailed, infoRules)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	gotValidationErrs, err := validator.VErrorsToMap(verrs, nil)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
	wantValidationErrs := validator.MapError{
		"Number": {"invalid number value"},
	}

	if !reflect.DeepEqual(wantValidationErrs, gotValidationErrs) {
		t.Fatalf("got %v, want %v", gotValidationErrs, wantValidationErrs)
	}

	// validate success

	infoSuccess := info{
		Number: 3,
	}

	verrs, err = validate.DoRules(infoSuccess, infoRules)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	if len(verrs) != 0 {
		t.Fatalf("shouldn't return any validation error: %v", verrs)
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

	validate := validator.New()

	verrs, err := validate.DoRules(infoFailed, infoRules)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	gotValidationErrs, err := validator.VErrorsToMap(verrs, nil)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
	wantValidationErrs := validator.MapError{
		"Gender": {"invalid gender value"},
	}

	if !reflect.DeepEqual(wantValidationErrs, gotValidationErrs) {
		t.Fatalf("got %v, want %v", gotValidationErrs, wantValidationErrs)
	}
}

func TestValidate_RegisterInclusionValidationParam_ParamIsEmpty(t *testing.T) {
	validate := validator.New()

	err := validate.RegisterInclusionValidationParam("", []string{"U", "M", "F"})
	if errors.Cause(err).Error() != "param can not be empty" {
		t.Fatal("should return `param can not be empty`")
	}
}

func TestValidate_RegisterInclusionValidationParamWithValidSliceTypeInvalid(t *testing.T) {
	validate := validator.New()

	err := validate.RegisterInclusionValidationParam("gender", "gender")
	if errors.Cause(err).Error() != "validSlice must be slice type" {
		t.Fatal("should return `validSlice must be slice type`")
	}
}

func TestVErrorsToMapWithDefaultTemplateMap(t *testing.T) {
	infoEmpty := info{}

	validate := validator.New()

	verrs, err := validate.DoRules(infoEmpty, infoRules)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	gotValidationErrs, err := validator.VErrorsToMap(verrs, nil)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
	wantValidationErrs := validator.MapError{
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
	templateMap := validator.TemplateMap{
		"lte":        "maximum length is {{.Param}}",
		"gte":        "minimum length is {{.Param}}",
		"zipcode_jp": "invalid zipcode format",
	}

	validate := validator.New()

	verrs, err := validate.DoRules(infoFailed, infoRules)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	gotValidationErrs, err := validator.VErrorsToMap(verrs, templateMap)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
	wantValidationErrs := validator.MapError{
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

	validate := validator.New()

	verrs, err := validate.DoRules(infoFailed, infoRules)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	gotValidationErrs, err := validator.VErrorsToMap(verrs, nil)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
	wantValidationErrs := validator.MapError{
		"Name": {"validation failed with lt=20"},
		"IPv4": {"validation failed with ipv4"},
	}

	if !reflect.DeepEqual(wantValidationErrs, gotValidationErrs) {
		t.Fatalf("got %v, want %v", gotValidationErrs, wantValidationErrs)
	}
}

func TestVErrorsToMapWithParseTemplateError(t *testing.T) {
	verrs := validator.Errors{
		{Field: "Name", Tag: "required"},
	}

	templateMap := validator.TemplateMap{
		"required": "maximum length is {{.Param",
	}
	_, err := validator.VErrorsToMap(verrs, templateMap)
	if err == nil {
		t.Fatal("should return err")
	}

	templateMap = validator.TemplateMap{
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

	validate := validator.New()

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
	wantValidationErrs := validator.MapError{
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

	validate := validator.New()

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
	wantValidationErrs := validator.MapError{
		"Phone": {"validation failed with phone"},
	}

	if !reflect.DeepEqual(wantValidationErrs, gotValidationErrs) {
		t.Fatalf("got %v, but want %v", gotValidationErrs, wantValidationErrs)
	}
}

func TestValidate_IsVar(t *testing.T) {
	validate := validator.New()

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

func TestValidate_DoRulesAndToMap(t *testing.T) {
	type info struct {
		Name string
	}

	infoRules := []validator.Rule{
		{Field: "Name", Tag: "required"},
	}

	infoEmpty := info{}

	validate := validator.New()

	validate.RegisterTemplateMap(validator.TemplateMap{
		"required": "custom require",
	})

	gotVerrMap, err := validate.DoRulesAndToMapError(infoEmpty, infoRules)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	wantVerrMap := validator.MapError{
		"Name": {"custom require"},
	}

	if !reflect.DeepEqual(gotVerrMap, wantVerrMap) {
		t.Fatalf("got %v, but want %v", gotVerrMap, wantVerrMap)
	}
}

func TestValidate_RegisterTemplateMapWithTemplateFailed(t *testing.T) {
	validate := validator.New()

	err := validate.RegisterTemplateMap(validator.TemplateMap{
		"required": "{{.invalidValue}}",
	})

	if err == nil {
		t.Fatal("should return err")
	}
}

func TestValidate_RegisterTemplateMapWithTagFailed(t *testing.T) {
	validate := validator.New()

	err := validate.RegisterTemplateMap(validator.TemplateMap{
		"": "{{template content}}",
	})

	if err == nil {
		t.Fatal("should return err")
	}
}

func TestValidate_DoRulesAndToMapWithTagName(t *testing.T) {
	type address struct {
		City    string `json:"json_city,omitempty"`
		Address string
	}

	type user struct {
		Name    string
		Age     int
		Address address `json:"address"`
	}

	userRules := []validator.Rule{
		{Field: "Address.City", Tag: "required"},
	}

	validate := validator.New()

	gotVerrMap, err := validate.DoRulesAndToMapErrorWithTagName(user{}, userRules, "json")
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	wantVerrMap := validator.MapError{
		"address.json_city": {"can not be blank"},
	}

	if !reflect.DeepEqual(gotVerrMap, wantVerrMap) {
		t.Fatalf("got %v, want %v", gotVerrMap, wantVerrMap)
	}
}

func TestVErrorMap_Error(t *testing.T) {
	var infoRules = []validator.Rule{
		{Field: "Name", Tag: "required"},
		{Field: "Password", Tag: "required"},
	}

	validate := validator.New()

	gotVMapErr, err := validate.DoRulesAndToMapError(info{}, infoRules)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	gotVMapErrString := gotVMapErr.Error()
	wantMapErrString1 := `Password:["can not be blank"] Name:["can not be blank"]`
	wantMapErrString2 := `Name:["can not be blank"] Password:["can not be blank"]`

	if !(gotVMapErrString == wantMapErrString1 || gotVMapErrString == wantMapErrString2) {
		t.Fatalf("snhould return `%v` or `%v`, \nbut return `%v`", wantMapErrString1, wantMapErrString2, gotVMapErr)
	}
}

func TestValidate_DoRulesWithEqfieldTag(t *testing.T) {
	type info struct {
		Password        string
		ConfirmPassword string
		OtherValidation string
	}

	var infoRules = []validator.Rule{
		{Field: "Password", Tag: "required,lte=2"},
		{Field: "ConfirmPassword", Tag: "required,lte=2,eqfield=Password"},
		{Field: "OtherValidation", Tag: "required"},
	}

	validate := validator.New()

	gotVerrMap, err := validate.DoRulesAndToMapError(info{Password: "1234", ConfirmPassword: "1111"}, infoRules)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	wantVerrMap := validator.MapError{
		"Password":        {"is too long, maximum length is 2"},
		"ConfirmPassword": {"validation failed with eqfield", "is too long, maximum length is 2"},
		"OtherValidation": {"can not be blank"},
	}

	if !reflect.DeepEqual(gotVerrMap, wantVerrMap) {
		t.Fatalf("got %v, want %v", gotVerrMap, wantVerrMap)
	}
}

func TestValidate_DoRulesToStruct(t *testing.T) {
	infoEmpty := info{}

	validate := validator.New()

	var infoErr *infoError
	validate.DoRulesToStruct(infoEmpty, infoRules, &infoErr)

	wantInfoErr := infoError{
		Name:     []string{"can not be blank"},
		Password: []string{"is too short, minimum length is 8"},
		Age:      []string{"is too small, minimum is 20"},
		Address:  []string{"can not be blank"},
		ZipCode:  []string{"invalid zipcode format, format is 123-1234"},
	}

	diff := testingutils.PrettyJsonDiff(wantInfoErr, infoErr)
	if len(diff) > 0 {
		t.Fatalf(diff)
	}
}

func TestValidate_DoRulesToStruct__toStructCheck1(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			if fmt.Sprint(r) != "toStruct must be pointer to pointer to struct" {
				t.Fatal("should panic 'toStruct must be pointer to pointer to struct'")
			}
		}
	}()

	infoEmpty := info{}

	validate := validator.New()

	var infoErr infoError
	validate.DoRulesToStruct(infoEmpty, infoRules, &infoErr)
}

func TestValidate_DoRulesToStruct__toStructCheck2(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			if fmt.Sprint(r) != "toStruct must be pointer to pointer to struct" {
				t.Fatal("should panic 'toStruct must be pointer to pointer to struct'")
			}
		}
	}()

	infoEmpty := info{}

	validate := validator.New()

	var infoErr *infoError
	validate.DoRulesToStruct(infoEmpty, infoRules, &infoErr)
}

func TestValidate_DoRulesToStruct__NoValidationErrors(t *testing.T) {
	infoEmpty := info{Name: "name", Password: "password", Age: 30, Address: "address", ZipCode: "000-0000"}

	validate := validator.New()

	var infoErr *infoError
	infoErr = &infoError{Password: []string{"for test"}}

	validate.DoRulesToStruct(infoEmpty, infoRules, &infoErr)
	if infoErr != nil {
		t.Fatal("infoErr should be nil")
	}
}
