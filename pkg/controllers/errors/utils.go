package errors

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
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
func ReportErrors(ctx context.Context, log logr.Logger, client ctrlclient.Client, err error) error {
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

		event := &corev1.Event{}
		event.GenerateName = "secret-rep-"
		event.InvolvedObject = involvedObject(intErr.Src)
		if intErr.Dst != nil {
			event.InvolvedObject = involvedObject(intErr.Dst)
		}
		event.Namespace = event.InvolvedObject.Namespace

		event.Reason = string(intErr.Reason)
		event.Message = intErr.Error()
		event.EventTime = metav1.NowMicro()
		event.ReportingController = "schrodit.tech/secret-replication-controller"
		event.ReportingInstance = "dummy"
		event.Action = "Reconcile"
		event.Type = "Warning"

		if err := client.Create(ctx, event); err != nil {
			reportErrs = append(reportErrs, intErr, err)
		}
	}

	return reportErrs.AggregateError()
}

func involvedObject(obj runtime.Object) corev1.ObjectReference {
	ref := corev1.ObjectReference{}
	ref.APIVersion = obj.GetObjectKind().GroupVersionKind().GroupVersion().String()
	ref.Kind = obj.GetObjectKind().GroupVersionKind().Kind

	if acc, ok := obj.(metav1.Object); ok {
		ref.Namespace = acc.GetNamespace()
		ref.Name = acc.GetName()
		ref.UID = acc.GetUID()
		ref.ResourceVersion = acc.GetResourceVersion()
	}
	return ref
}
