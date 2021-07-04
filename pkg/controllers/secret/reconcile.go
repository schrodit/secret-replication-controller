package secretctrl

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	"github.com/schrodit/secret-replication-controller/pkg/apis/core/v1alpha1"
	"github.com/schrodit/secret-replication-controller/pkg/apis/core/v1alpha1/helper"
	"github.com/schrodit/secret-replication-controller/pkg/controllers/errors"
	interrors "github.com/schrodit/secret-replication-controller/pkg/controllers/errors"
	"github.com/schrodit/secret-replication-controller/pkg/replicator"
)

func (c *secretController) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	ctx = logr.NewContext(ctx, c.log.WithValues("name", req.Name, "namespace", req.Namespace))
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
	log := logr.FromContextOrDiscard(ctx)
	log.V(10).Info("check replication for secret")
	namespacesVal, hasNamespacesAnn := helper.GetAnnotation(secret, v1alpha1.SecretReplicationNamespacesAnnotations)
	_, hasAllNamespacesAnn := helper.GetAnnotation(secret, v1alpha1.SecretReplicationAllNamespacesAnnotations)
	if !hasNamespacesAnn && !hasAllNamespacesAnn {
		log.V(10).Info("secret not applicable for replication")
		return nil
	}

	var namespaces []string

	if hasAllNamespacesAnn {
		var err error
		namespaces, err = c.getAllNamespaces(ctx, secret)
		if err != nil {
			return err
		}
	} else if hasNamespacesAnn {
		var err error
		namespaces, err = c.parseNamespaces(ctx, secret, namespacesVal)
		if err != nil {
			return c.Report(ctx, err)
		}
	}

	var (
		allErrs    = interrors.ErrorList{}
		replicator = replicator.New(c.client, secret)
	)
	for _, namespace := range namespaces {

		// ensure the secret for a namespace
		if err := replicator.ReplicateTo(ctx, namespace); err != nil {
			allErrs = append(allErrs, err)
		}
	}

	return c.Report(ctx, allErrs)
}

func (c *secretController) parseNamespaces(ctx context.Context, secret *corev1.Secret, namespaces string) ([]string, error) {
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

func (c *secretController) getAllNamespaces(ctx context.Context, secret *corev1.Secret) ([]string, error) {
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
