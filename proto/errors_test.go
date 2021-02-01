package proto

import (
	"testing"

	"github.com/theplant/testingutils/fatalassert"
)

func TestErrorMessage(t *testing.T) {
	tests := []struct {
		name  string
		error Error

		expectedValidationError Error
	}{
		{
			name: "one FieldViolations and top code has value",

			error: Error{
				Code: "top code",
				FieldViolations: []*ValidationError_FieldViolation{
					{
						Field: "field1",
						Code:  "field1 code",
						Msg:   "field1 message",
					},
				},
			},

			expectedValidationError: Error{
				Code: "top code",
				FieldViolations: []*ValidationError_FieldViolation{
					{
						Field: "field1",
						Code:  "field1 code",
						Msg:   "field1 message",
					},
				},
			},
		},

		{
			name: "one FieldViolations and top code has not value",

			error: Error{
				FieldViolations: []*ValidationError_FieldViolation{
					{
						Field: "field1",
						Code:  "field1 code",
						Msg:   "field1 message",
						Hmsg:  "field1 human message",
					},
				},
			},

			expectedValidationError: Error{
				Code: "field1 code",
				Hmsg: "field1 human message",
				FieldViolations: []*ValidationError_FieldViolation{
					{
						Field: "field1",
						Code:  "field1 code",
						Msg:   "field1 message",
						Hmsg:  "field1 human message",
					},
				},
			},
		},

		{
			name: "multiple FieldViolations and top code has value",

			error: Error{
				Code: "top code",
				FieldViolations: []*ValidationError_FieldViolation{
					{
						Field: "field1",
						Code:  "field1 code",
						Msg:   "field1 message",
					},
					{
						Field: "field2",
						Code:  "field2 code",
						Msg:   "field2 message",
					},
				},
			},

			expectedValidationError: Error{
				Code: "top code",
				FieldViolations: []*ValidationError_FieldViolation{
					{
						Field: "field1",
						Code:  "field1 code",
						Msg:   "field1 message",
					},
					{
						Field: "field2",
						Code:  "field2 code",
						Msg:   "field2 message",
					},
				},
			},
		},

		{
			name: "multiple FieldViolations and top code has not value",

			error: Error{
				FieldViolations: []*ValidationError_FieldViolation{
					{
						Field: "field1",
						Code:  "field1 code",
						Msg:   "field1 message",
					},
					{
						Field: "field2",
						Code:  "field2 code",
						Msg:   "field2 message",
					},
				},
			},

			expectedValidationError: Error{
				FieldViolations: []*ValidationError_FieldViolation{
					{
						Field: "field1",
						Code:  "field1 code",
						Msg:   "field1 message",
					},
					{
						Field: "field2",
						Code:  "field2 code",
						Msg:   "field2 message",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotValidationError := Error(*test.error.Message().(*ValidationError))
			fatalassert.Equal(t, test.expectedValidationError, gotValidationError)
		})
	}
}
