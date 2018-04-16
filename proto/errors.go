package proto

import (
	"net/http"

	"github.com/golang/protobuf/proto"
)

type Error ValidationError

func (err *Error) Message() proto.Message {
	v := ValidationError(*err)

	if v.Code == "" && len(v.FieldViolations) == 1 {
		v.Code = v.FieldViolations[0].Code
	}

	return &v
}

func (err *Error) HTTPStatusCode() int {
	return http.StatusUnprocessableEntity // 422
}

func (err *Error) Error() string {
	return "validation error"
}
