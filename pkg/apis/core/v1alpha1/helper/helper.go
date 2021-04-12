package helper

import (
	"github.com/schrodit/secret-replication-controller/pkg/apis/core/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetAnnotation returns the value of one of the matching annotations.
// Returns false if no annotations matches.
func GetAnnotation(obj client.Object, annotations *v1alpha1.AnnotationSet) (string, bool) {
	return annotations.Get(obj.GetAnnotations())
}
