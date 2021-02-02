package proto

import (
	"net/http"

	"github.com/golang/protobuf/proto"
)

type Error ValidationError

func (err *Error) Message() proto.Message {
	v := ValidationError(*err)

	if len(v.FieldViolations) == 1 {
		f0 := v.FieldViolations[0]
		if v.Code == "" {
			v.Code = f0.Code
		}
		if v.DefaultViewMsg == "" {
			v.DefaultViewMsg = f0.DefaultViewMsg
		}
	}

	return &v
}

func (err *Error) HTTPStatusCode() int {
	return http.StatusUnprocessableEntity // 422
}

func (err *Error) Error() string {
	return "validation error"
}
