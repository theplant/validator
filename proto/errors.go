package proto

import (
	"net/http"

	"github.com/gogo/protobuf/proto"
)

type Error struct {
	ValidationError
}

func (err Error) Message() proto.Message {
	return &err
}

func (err Error) HTTPStatusCode() int {
	return http.StatusUnprocessableEntity // 422
}

func (err Error) Error() string {
	return "validation error"
}
