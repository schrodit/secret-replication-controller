package ingressctrl

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	"github.com/schrodit/secret-replication-controller/pkg/apis/core/v1alpha1"
	"github.com/schrodit/secret-replication-controller/pkg/apis/core/v1alpha1/helper"
	interrors "github.com/schrodit/secret-replication-controller/pkg/controllers/errors"
	"github.com/schrodit/secret-replication-controller/pkg/replicator"
)

func (c *IngressController) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	ctx = logr.NewContext(ctx, c.log.WithValues("name", req.Name, "namespace", req.Namespace))
	ingress := &networkingv1beta1.Ingress{}
	if err := c.client.Get(ctx, req.NamespacedName, ingress); err != nil {
		return reconcile.Result{}, err
	}

	if err := c.reconcile(ctx, ingress); err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (c *IngressController) reconcile(ctx context.Context, ingress *networkingv1beta1.Ingress) error {
	log := logr.FromContextOrDiscard(ctx)
	log.V(10).Info("check replication for ingress")
	srcNamespace, ok := helper.GetAnnotation(ingress, v1alpha1.SecretReplicationFromNamespaceAnnotations)
	if !ok {
		log.V(10).Info("ingress not applicable for replication")
		return nil
	}

	// get all secrets from the ingress
	usedSecrets := getSecretsFromIngress(ingress)
	if len(usedSecrets) == 0 {
		log.V(10).Info("no secrets used by ingress")
		return nil
	}

	// check if defined namespace exists
	if err := c.client.Get(ctx, client.ObjectKey{Name: srcNamespace}, &corev1.Namespace{}); err != nil {
		return c.Report(ctx, fmt.Errorf("namespace %q not found", srcNamespace))
	}
	targetNamespace := ingress.Namespace

	allErrs := interrors.ErrorList{}
	for _, secretName := range usedSecrets {

		// only sync secrets that exist in the given namespace
		secret := &corev1.Secret{}
		if err := c.client.Get(ctx, client.ObjectKey{Name: secretName, Namespace: srcNamespace}, secret); err != nil {
			allErrs = append(allErrs, fmt.Errorf("unable to find secret %q: %w", secretName, err))
			continue
		}

		if err := replicator.New(c.client, secret).ReplicateTo(ctx, targetNamespace); err != nil {
			allErrs = append(allErrs, err)
		}
	}

	return c.Report(ctx, allErrs)
}

// getSecretsFromIngress returns all used secrets for the ingress.
func getSecretsFromIngress(ingress *networkingv1beta1.Ingress) []string {
	secrets := sets.NewString()
	for _, tls := range ingress.Spec.TLS {
		secrets.Insert(tls.SecretName)
	}
	return secrets.List()
}
