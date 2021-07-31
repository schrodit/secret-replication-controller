package v1alpha1

import "k8s.io/apimachinery/pkg/util/sets"

// User facing annotations
const (
	// DefaultPrefix defines the default prefix for user facing annotations
	DefaultAnnotationPrefix = "replication.schrodit.tech"

	// NamespacesAnnotation is the name of the annotation that defines the namespaces where the annotated resource should be replicated to.
	NamespacesAnnotation = "namespaces"

	// SecretReplicationNamespacesAnnotation is the name of the annotation that defines the namespaces where the annotated resource should be replicated to.
	SecretReplicationNamespacesAnnotation = "replication.schrodit.tech/namespaces"

	// AllNamespacesAnnotation is the name of the annotation that defines that the annotated resource should be replicated to all namespaces.
	AllNamespacesAnnotation = "all"

	// SecretReplicationAllNamespacesAnnotation is the name of the annotation that defines that the annotated resource should be replicated to all namespaces.
	SecretReplicationAllNamespacesAnnotation = "replication.schrodit.tech/all"

	// FromNamespaceAnnotation is the name of the annotation that defines where the defined secret of the ingress should be synced from.
	FromNamespaceAnnotation = "from-namespace"

	// SecretReplicationFromNamespaceAnnotation is the name of the annotation that defines where the defined secret of the ingress should be synced from.
	SecretReplicationFromNamespaceAnnotation = "replication.schrodit.tech/from-namespace"
)

var (
	// SecretReplicationNamespacesAnnotations are the names of the annotation that defines the namespaces where the annotated resource should be replicated to.
	SecretReplicationNamespacesAnnotations = NewAnnotationSet(NamespacesAnnotation, DefaultAnnotationPrefix)

	// SecretReplicationAllNamespacesAnnotations are the names of the annotation that defines that the annotated resource should be replicated to all namespaces.
	SecretReplicationAllNamespacesAnnotations = NewAnnotationSet(AllNamespacesAnnotation, DefaultAnnotationPrefix)

	// SecretReplicationFromNamespaceAnnotations are the names of the annotation that defines where the defined secret of the ingress should be synced from.
	SecretReplicationFromNamespaceAnnotations = NewAnnotationSet(FromNamespaceAnnotation, DefaultAnnotationPrefix)
)

// SecretReplicationReplicaOfAnnotation is the name of the annotation that defines the source resource of the current resource.
const SecretReplicationReplicaOfAnnotation = "replication.schrodit.tech/replicaOf"

// SecretReplicationLastHashAnnotation is the name of the annotation that defines the last observed hash of the replicating secret.
const SecretReplicationLastObservedHashAnnotation = "replication.schrodit.tech/lastObservedHash"

const Separator = "/"

type AnnotationSet struct {
	Default string
	Key     string
	set     sets.String
}

// NewAnnotationSet creates a new annotation set.
func NewAnnotationSet(key string, defaultAnn string) *AnnotationSet {
	set := &AnnotationSet{
		Default: defaultAnn,
		Key:     key,
		set:     sets.NewString(),
	}
	set.Add(defaultAnn)
	return set
}

// Add adds multiple annotation prefixes.
func (s AnnotationSet) Add(prefix string) string {
	ann := s.makeAnnotation(prefix)
	s.set.Insert(ann)
	return ann
}

func (s AnnotationSet) makeAnnotation(prefix string) string {
	return prefix + Separator + s.Key
}

// Get returns the value of one of the matching annotations.
// Returns false if no annotations matches.
func (s AnnotationSet) Get(annotations map[string]string) (string, bool) {
	if len(annotations) == 0 {
		return "", false
	}

	for _, ann := range s.set.List() {
		if val, ok := annotations[ann]; ok {
			return val, ok
		}
	}
	return "", false
}

// Reset resets to the default annotation
func (s *AnnotationSet) Reset() {
	s.set = sets.NewString(s.makeAnnotation(s.Default))
}

// List returns all annotations as list.
func (s AnnotationSet) List() []string {
	return s.set.List()
}
