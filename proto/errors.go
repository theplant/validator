package proto

import (
	"net/http"

	"github.com/gogo/protobuf/proto"
)

type Error ValidationError

func (err *Error) Message() proto.Message {
	v := ValidationError(*err)
	return &v
}

func (err *Error) HTTPStatusCode() int {
	return http.StatusUnprocessableEntity // 422
}

func (err *Error) Error() string {
	return "validation error"
}
