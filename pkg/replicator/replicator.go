package replicator

import (
	"context"
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/schrodit/secret-replication-controller/pkg/apis/core/v1alpha1"
	"github.com/schrodit/secret-replication-controller/pkg/controllers/errors"
	"k8s.io/apimachinery/pkg/types"
)

type Replicator struct {
	client client.Client
	secret *corev1.Secret
}

func New(kubeClient client.Client, secret *corev1.Secret) *Replicator {
	return &Replicator{
		client: kubeClient,
		secret: secret,
	}
}

// ReplicateTo replicates the secret to the given namespace.
func (r *Replicator) ReplicateTo(ctx context.Context, namespace string) error {
	key := types.NamespacedName{
		Name:      r.secret.Name,
		Namespace: namespace,
	}

	// check if secret is already created
	repSecret := &corev1.Secret{}
	if err := r.client.Get(ctx, key, repSecret); err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Error{
				Src:    r.secret,
				Reason: errors.InternalError,
				Msg:    "unable to get secret to create",
				Err:    err,
			}
		}

		// secret is not created yet so lets create it
		repSecret := &corev1.Secret{}
		repSecret.Name = key.Name
		repSecret.Namespace = key.Namespace
		repSecret.Data = r.secret.Data
		repSecret.Annotations = map[string]string{
			v1alpha1.SecretReplicationLasObservedGenerationAnnotation: strconv.Itoa(int(r.secret.Generation)),
			v1alpha1.SecretReplicationReplicaOfAnnotation:             types.NamespacedName{Name: r.secret.Name, Namespace: r.secret.Namespace}.String(),
		}

		if err := r.client.Create(ctx, repSecret); err != nil {
			return errors.Error{
				Src:    r.secret,
				Reason: errors.CreateError,
				Msg:    fmt.Sprintf("unable to create replicated secret in namespace %s", key.Namespace),
				Err:    err,
			}
		}
		return nil
	}

	if !IsApplicableForUpdate(r.secret, repSecret, false) {
		return nil
	}

	repSecret.Data = r.secret.Data
	repSecret.Annotations = map[string]string{
		v1alpha1.SecretReplicationLasObservedGenerationAnnotation: strconv.Itoa(int(r.secret.Generation)),
	}

	if err := r.client.Update(ctx, repSecret); err != nil {
		return errors.Error{
			Src:    r.secret,
			Dst:    repSecret,
			Reason: errors.UpdateError,
			Msg:    fmt.Sprintf("unable to update replicated secret in namespace %s", key.Namespace),
			Err:    err,
		}
	}
	return nil
}

// IsApplicableForUpdate checks whether a resource is applicable for an update.
// The destination resource is not overwritten if the replicaOf annotation does not matches the source resource.
// Setting force also updates the destination resource if no replicaOf is set.
func IsApplicableForUpdate(src, dst metav1.Object, force bool) bool {
	// only update if the observed generation differ
	if dst.GetAnnotations()[v1alpha1.SecretReplicationLasObservedGenerationAnnotation] == strconv.Itoa(int(src.GetGeneration())) {
		return false
	}

	// do not update if the secret is not controlled by the current secret
	replicaOf, ok := dst.GetAnnotations()[v1alpha1.SecretReplicationReplicaOfAnnotation]
	if !ok && force {
		return true
	}

	return types.NamespacedName{Name: src.GetName(), Namespace: src.GetNamespace()}.String() == replicaOf
}
