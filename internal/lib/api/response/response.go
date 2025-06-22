package response

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Response struct {
	Status TaskStatus `json:"status"`
	Error  string     `json:"error,omitempty"`
}

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
	StatusCancelled TaskStatus = "cancelled"
	StatusDeleted   TaskStatus = "deleted"
)

func Error(msg string) Response {
	return Response{
		Status: StatusFailed,
		Error:  msg,
	}
}

func ValidationError(errs validator.ValidationErrors) Response {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field", err.Field()))
		case "min":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s must be at least %s characters long", err.Field(), err.Param()))
		case "max":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s must be at most %s characters long", err.Field(), err.Param()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}

	return Response{
		Status: StatusFailed,
		Error:  strings.Join(errMsgs, ", "),
	}
}

func Success() Response {
	return Response{
		Status: StatusPending,
	}
}

func Status(status TaskStatus) Response {
	return Response{
		Status: status,
	}
}

func Deleted() Response {
	return Response{
		Status: StatusDeleted,
	}
}
