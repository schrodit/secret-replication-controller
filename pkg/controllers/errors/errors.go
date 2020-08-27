package errors

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
)

// ErrorList defines a list of errors
type ErrorList []error

var _ error = ErrorList{}

func (l ErrorList) Error() string {
	err := AggregateMultiError(l)
	if err != nil {
		return ""
	}
	return err.Error()
}

// AggregateError aggregates multiple errors to one error
func (l ErrorList) AggregateError() error {
	return AggregateMultiError(l)
}

// Reason defines a specific error reason
type Reason string

// Error a custom replication error.
type Error struct {
	// Src defines the source resource of the error.
	Src runtime.Object
	// Dst defines the destinition resource where the error occurred
	Dst runtime.Object
	// Reason defines a specific error reason
	Reason Reason
	// Msg defines a optional detailed error message
	Msg string
	// Err defines a optional wrapped error
	Err error
}

var _ error = Error{}

func (e Error) Error() string {
	errMsg := e.Msg
	if e.Err != nil {
		errMsg = fmt.Sprintf("%s: %s", errMsg, e.Err.Error())
	}
	return errMsg
}

func (e Error) Unwrap() error {
	return e.Err
}
