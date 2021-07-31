package ingressctrl_test

import (
	"context"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/schrodit/secret-replication-controller/pkg/apis/core/v1alpha1"
	ingressctrl "github.com/schrodit/secret-replication-controller/pkg/controllers/ingress"
	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("controller", func() {

	var (
		ctrl       reconcile.Reconciler
		secret     *corev1.Secret
		namespaces []string
	)

	BeforeEach(func() {
		secret = &corev1.Secret{}
		secret.GenerateName = "e2e-"
		secret.Namespace = "default"
		secret.Data = map[string][]byte{
			"key": []byte("value"),
		}

		Expect(client.Create(context.TODO(), secret)).To(Succeed())
		namespaces = make([]string, 0)

		ctrl = ingressctrl.New(logr.Discard(), client, record.NewFakeRecorder(1024))
	})

	AfterEach(func() {
		ctx := context.Background()
		defer ctx.Done()
		Expect(client.Delete(ctx, secret)).To(Succeed())

		for _, ns := range namespaces {
			namespace := &corev1.Namespace{}
			namespace.Name = ns
			Expect(client.Delete(ctx, namespace)).To(Succeed())
		}
	})

	It("should create a replicated secret in one namespace", func() {
		ctx := context.Background()
		defer ctx.Done()

		ns := &corev1.Namespace{}
		ns.GenerateName = "e2e-"
		Expect(client.Create(ctx, ns))
		namespaces = append(namespaces, ns.Name)

		ingress := defaultIngress()
		ingress.GenerateName = "e2e-"
		ingress.Namespace = ns.Name
		ingress.Annotations = map[string]string{
			v1alpha1.SecretReplicationFromNamespaceAnnotation: secret.Namespace,
		}
		ingress.Spec.TLS = []networkingv1beta1.IngressTLS{
			{
				Hosts:      []string{"example.com"},
				SecretName: secret.Name,
			},
		}

		Expect(client.Create(ctx, ingress)).To(Succeed())

		_, err := ctrl.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: ingress.Name, Namespace: ingress.Namespace}})
		Expect(err).ToNot(HaveOccurred())

		newSecret := &corev1.Secret{}
		Expect(client.Get(ctx, types.NamespacedName{Name: secret.Name, Namespace: ns.Name}, newSecret)).To(Succeed())

		Expect(newSecret.Data).To(Equal(secret.Data))
		Expect(newSecret.Annotations).To(HaveKey(v1alpha1.SecretReplicationLastObservedHashAnnotation))
		Expect(newSecret.Annotations[v1alpha1.SecretReplicationLastObservedHashAnnotation]).ToNot(Equal(""))
	})

	It("should create a replicated secret in one namespace with a custom prefix", func() {
		ctx := context.Background()
		defer ctx.Done()

		customPrefix := "some-prefix"
		v1alpha1.SecretReplicationFromNamespaceAnnotations.Add(customPrefix)
		defer func() {
			v1alpha1.SecretReplicationFromNamespaceAnnotations.Reset()
		}()

		ns := &corev1.Namespace{}
		ns.GenerateName = "e2e-"
		Expect(client.Create(ctx, ns))
		namespaces = append(namespaces, ns.Name)

		ingress := defaultIngress()
		ingress.GenerateName = "e2e-"
		ingress.Namespace = ns.Name
		ingress.Annotations = map[string]string{
			"some-prefix/from-namespace": secret.Namespace,
		}
		ingress.Spec.TLS = []networkingv1beta1.IngressTLS{
			{
				Hosts:      []string{"example.com"},
				SecretName: secret.Name,
			},
		}

		Expect(client.Create(ctx, ingress)).To(Succeed())

		_, err := ctrl.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: ingress.Name, Namespace: ingress.Namespace}})
		Expect(err).ToNot(HaveOccurred())

		newSecret := &corev1.Secret{}
		Expect(client.Get(ctx, types.NamespacedName{Name: secret.Name, Namespace: ns.Name}, newSecret)).To(Succeed())

		Expect(newSecret.Data).To(Equal(secret.Data))
		Expect(newSecret.Annotations).To(HaveKey(v1alpha1.SecretReplicationLastObservedHashAnnotation))
		Expect(newSecret.Annotations[v1alpha1.SecretReplicationLastObservedHashAnnotation]).ToNot(Equal(""))
	})

	It("should update an existing secret when data of the source is updated", func() {
		ctx := context.Background()
		defer ctx.Done()

		ns := &corev1.Namespace{}
		ns.GenerateName = "e2e-"
		Expect(client.Create(ctx, ns))
		namespaces = append(namespaces, ns.Name)

		ingress := defaultIngress()
		ingress.GenerateName = "e2e-"
		ingress.Namespace = ns.Name
		ingress.Annotations = map[string]string{
			v1alpha1.SecretReplicationFromNamespaceAnnotation: secret.Namespace,
		}
		ingress.Spec.TLS = []networkingv1beta1.IngressTLS{
			{
				Hosts:      []string{"example.com"},
				SecretName: secret.Name,
			},
		}

		Expect(client.Create(ctx, ingress)).To(Succeed())

		_, err := ctrl.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: ingress.Name, Namespace: ingress.Namespace}})
		Expect(err).ToNot(HaveOccurred())

		newSecret := &corev1.Secret{}
		Expect(client.Get(ctx, types.NamespacedName{Name: secret.Name, Namespace: ns.Name}, newSecret)).To(Succeed())

		secret.Data = map[string][]byte{
			"other": []byte("test"),
		}
		Expect(client.Update(ctx, secret)).To(Succeed())

		_, err = ctrl.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: ingress.Name, Namespace: ingress.Namespace}})
		Expect(err).ToNot(HaveOccurred())

		Expect(client.Get(ctx, types.NamespacedName{Name: secret.Name, Namespace: ns.Name}, newSecret)).To(Succeed())
		Expect(newSecret.Data).To(Equal(secret.Data))
	})

})

func defaultIngress() *networkingv1beta1.Ingress {
	ingress := &networkingv1beta1.Ingress{}
	ingress.Spec.Rules = []networkingv1beta1.IngressRule{
		{},
	}
	return ingress
}
