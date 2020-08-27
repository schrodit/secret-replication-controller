package controller

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/schrodit/secret-replication-controller/pkg/apis/core/v1alpha1"
	"github.com/schrodit/secret-replication-controller/pkg/controllers/errors"
	interrors "github.com/schrodit/secret-replication-controller/pkg/controllers/errors"
)

func (c *secretController) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()
	defer ctx.Done()
	secret := &corev1.Secret{}
	if err := c.client.Get(ctx, req.NamespacedName, secret); err != nil {
		return reconcile.Result{}, err
	}

	if err := c.reconcile(ctx, secret); err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (c *secretController) reconcile(ctx context.Context, secret *corev1.Secret) error {
	c.log.V(10).Info("check replication for secret", "name", secret.Name, "namespace", secret.Namespace)
	if !metav1.HasAnnotation(secret.ObjectMeta, v1alpha1.SecretReplicationNamespacesAnnotation) && !metav1.HasAnnotation(secret.ObjectMeta, v1alpha1.SecretReplicationAllNamespacesAnnotation) {
		c.log.V(10).Info("secret not applicable for replication", "name", secret.Name, "namespace", secret.Namespace)
		return nil
	}

	namespaces, err := c.getNamespaces(ctx, secret)
	if err != nil {
		return errors.ReportErrors(ctx, c.log, c.client, err)
	}

	allErrs := interrors.ErrorList{}
	for _, namespace := range namespaces {

		// ensure the secret for a namespace
		if err := c.ensureSecret(ctx, secret, namespace); err != nil {
			allErrs = append(allErrs, err)
		}

	}

	return errors.ReportErrors(ctx, c.log, c.client, allErrs)
}

func (c *secretController) getNamespaces(ctx context.Context, secret *corev1.Secret) ([]string, error) {
	// split the given namespaces defined in the annotation
	if namespaces, ok := secret.Annotations[v1alpha1.SecretReplicationNamespacesAnnotation]; ok {
		namespaceList := strings.Split(namespaces, ",")

		// lets validate here if the namespace exists
		for _, nsName := range namespaceList {
			ns := &corev1.Namespace{}
			if err := c.client.Get(ctx, types.NamespacedName{Name: nsName}, ns); err != nil {
				return nil, errors.Error{
					Src:    secret,
					Reason: errors.InvalidNamespace,
					Err:    err,
				}
			}
			if !ns.DeletionTimestamp.IsZero() {
				return nil, errors.Error{
					Src:    secret,
					Reason: errors.InvalidNamespace,
					Msg:    fmt.Sprintf("namespace %s is marked for deletion", nsName),
				}
			}
		}

		return namespaceList, nil
	}

	// ignore the specific namespaces if the secret should be replicated to all
	if secret.Annotations[v1alpha1.SecretReplicationAllNamespacesAnnotation] == "true" {
		nsList := &corev1.NamespaceList{}
		if err := c.client.List(ctx, nsList); err != nil {
			return nil, errors.Error{
				Src:    secret,
				Reason: errors.InvalidNamespace,
				Msg:    "unable to list all namespaces",
				Err:    err,
			}
		}

		namespaces := make([]string, len(nsList.Items))
		for i, ns := range nsList.Items {
			namespaces[i] = ns.Name
		}

		return namespaces, nil
	}

	return nil, nil
}

func (c *secretController) ensureSecret(ctx context.Context, secret *corev1.Secret, namespace string) error {
	key := types.NamespacedName{
		Name:      secret.Name,
		Namespace: namespace,
	}

	// check if secret is already created
	repSecret := &corev1.Secret{}
	if err := c.client.Get(ctx, key, repSecret); err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Error{
				Src:    secret,
				Reason: errors.InternalError,
				Msg:    "unable to get secret to create",
				Err:    err,
			}
		}

		// secret is not created yet so lets create it
		repSecret := &corev1.Secret{}
		repSecret.Name = key.Name
		repSecret.Namespace = key.Namespace
		repSecret.Data = secret.Data
		repSecret.Annotations = map[string]string{
			v1alpha1.SecretReplicationLasObservedGenerationAnnotation: strconv.Itoa(int(secret.Generation)),
			v1alpha1.SecretReplicationReplicaOfAnnotation:             types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}.String(),
		}

		if err := c.client.Create(ctx, repSecret); err != nil {
			return errors.Error{
				Src:    secret,
				Reason: errors.CreateError,
				Msg:    fmt.Sprintf("unable to create replicated secret in namespace %s", key.Namespace),
				Err:    err,
			}
		}
		return nil
	}

	if !IsApplicableForUpdate(secret, repSecret, false) {
		return nil
	}

	repSecret.Data = secret.Data
	repSecret.Annotations = map[string]string{
		v1alpha1.SecretReplicationLasObservedGenerationAnnotation: strconv.Itoa(int(secret.Generation)),
	}

	if err := c.client.Update(ctx, repSecret); err != nil {
		return errors.Error{
			Src:    secret,
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
