package replicator

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
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
	log := logr.FromContextOrDiscard(ctx)
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
		log.V(3).Info("Secret in target namespace not found. Creating...", "target", namespace)

		srcHash, err := secretHash(r.secret)
		if err != nil {
			return fmt.Errorf("unable to hash data of source secret: %w", err)
		}

		// secret is not created yet so lets create it
		repSecret := &corev1.Secret{}
		repSecret.Name = key.Name
		repSecret.Namespace = key.Namespace
		repSecret.Data = r.secret.Data
		repSecret.Annotations = map[string]string{
			v1alpha1.SecretReplicationLastObservedHashAnnotation: srcHash,
			v1alpha1.SecretReplicationReplicaOfAnnotation:        types.NamespacedName{Name: r.secret.Name, Namespace: r.secret.Namespace}.String(),
		}

		if err := r.client.Create(ctx, repSecret); err != nil {
			return errors.Error{
				Src:    r.secret,
				Reason: errors.CreateError,
				Msg:    fmt.Sprintf("unable to create replicated secret in namespace %s: %s", key.Namespace, err.Error()),
				Err:    err,
			}
		}
		return nil
	}

	update, srcHash, err := IsApplicableForUpdate(r.secret, repSecret, false)
	if err != nil {
		return err
	}
	if !update {
		return nil
	}
	log.V(3).Info("Secret out-of-date. Updating...")

	repSecret.Data = r.secret.Data
	metav1.SetMetaDataAnnotation(&repSecret.ObjectMeta, v1alpha1.SecretReplicationLastObservedHashAnnotation, srcHash)

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
// The function returns if the secret is applicated to be updasted, the new src hash and a optional error.
// The source hash is only returned if the secret should be updated.
func IsApplicableForUpdate(src, dst *corev1.Secret, force bool) (bool, string, error) {
	lastObservedHash := dst.Annotations[v1alpha1.SecretReplicationLastObservedHashAnnotation]

	srcHash, err := secretHash(src)
	if err != nil {
		return false, "", fmt.Errorf("unable to hash data of source secret: %w", err)
	}

	// only update if the observed hash differ
	if lastObservedHash == srcHash {
		return false, "", nil
	}

	// do not update if the secret is not controlled by the current secret
	replicaOf, ok := dst.Annotations[v1alpha1.SecretReplicationReplicaOfAnnotation]
	if !ok && force {
		return true, srcHash, nil
	}

	return types.NamespacedName{Name: src.GetName(), Namespace: src.GetNamespace()}.String() == replicaOf, srcHash, nil
}

// secretHash creates a hash value of the data of the given secret.
func secretHash(secret *corev1.Secret) (string, error) {
	// create a hashable representation of the data using json
	data, err := json.Marshal(secret.Data)
	if err != nil {
		return "", err
	}

	h := sha1.New()
	_, _ = h.Write(data)
	return hex.EncodeToString(h.Sum(nil)), nil
}
