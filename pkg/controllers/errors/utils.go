package errors

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

// AggregateMultiError aggregates multiple errors.
func AggregateMultiError(errs []error) error {
	if len(errs) == 0 {
		return nil
	}

	errMsg := ""
	for _, err := range errs {
		errMsg = fmt.Sprintf("%s - %s", errMsg, err.Error())
	}

	return errors.New(errMsg)
}

// ReportErrors reports all errors of a known internal type as events.
// unknown errors are logged.
func ReportErrors(ctx context.Context, log logr.Logger, eventRecorder record.EventRecorder, err error) error {
	allErrs := ErrorList{err}
	if errs, ok := err.(ErrorList); ok {
		allErrs = errs
	}

	reportErrs := ErrorList{}
	for _, err := range allErrs {
		// todo: support wrapped errors
		intErr, ok := err.(Error)
		if !ok {
			reportErrs = append(reportErrs, err)
			log.Error(err, "")
			continue
		}
		eventRecorder.Event(intErr.Src, corev1.EventTypeWarning, string(intErr.Reason), intErr.Msg)
		if intErr.Dst != nil {
			eventRecorder.Event(intErr.Dst, corev1.EventTypeWarning, string(intErr.Reason), intErr.Msg)
		}
	}

	return reportErrs.AggregateError()
}

// ErrorReporter is a struct that reports aggreagted errors.
// Is basically a simple wrapper for ReportErrors.
type ErrorReporter struct {
	recorder record.EventRecorder
}

// NewErrorReporter creates a new error reporter
func NewErrorReporter(recorder record.EventRecorder) *ErrorReporter {
	return &ErrorReporter{
		recorder: recorder,
	}
}

func (er ErrorReporter) Report(ctx context.Context, err error) error {
	log := logr.FromContextOrDiscard(ctx)
	return ReportErrors(ctx, log, er.recorder, err)
}
