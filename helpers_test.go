package validator_test

import (
	"fmt"
	"testing"

	"github.com/theplant/validator"
)

var isStructZeroCases = []struct {
	info           *info
	expectedIsZero bool
}{
	{
		info:           &info{},
		expectedIsZero: true,
	},

	{
		info:           &info{Name: "", Password: ""},
		expectedIsZero: true,
	},

	{
		info:           &info{Name: "name"},
		expectedIsZero: false,
	},

	{
		info:           &info{Name: "name", Password: "password"},
		expectedIsZero: false,
	},
}

func TestIsStructZero(t *testing.T) {
	for _, c := range isStructZeroCases {
		isZero := validator.IsStructZero(c.info)
		if c.expectedIsZero != isZero {
			t.Errorf("expected: %v, but got: %v", c.expectedIsZero, isZero)
		}
	}
}

func TestIsStructZero__sIfaceMustBeStruct(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			if fmt.Sprint(r) != "sIface must be pointer to struct" {
				t.Fatal("sIface must be pointer to struct")
			}
		}
	}()

	validator.IsStructZero(info{})
}
