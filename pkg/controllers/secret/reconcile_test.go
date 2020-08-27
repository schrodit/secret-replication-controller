package controller

import (
	"context"
	"fmt"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/schrodit/secret-replication-controller/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("controller", func() {

	var (
		ctrl       *secretController
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

		ctrl = &secretController{
			client: client,
		}
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

		By("create test namespace")
		ns := &corev1.Namespace{}
		ns.GenerateName = "e2e-"
		Expect(client.Create(ctx, ns))
		namespaces = append(namespaces, ns.Name)

		secret.Annotations = map[string]string{
			v1alpha1.SecretReplicationNamespacesAnnotation: ns.Name,
		}
		Expect(client.Update(ctx, secret)).To(Succeed())

		_, err := ctrl.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}})
		Expect(err).ToNot(HaveOccurred())

		newSecret := &corev1.Secret{}
		Expect(client.Get(ctx, types.NamespacedName{Name: secret.Name, Namespace: ns.Name}, newSecret)).To(Succeed())

		Expect(newSecret.Data).To(Equal(secret.Data))
		Expect(newSecret.Annotations).To(HaveKeyWithValue(v1alpha1.SecretReplicationLasObservedGenerationAnnotation, strconv.Itoa(int(secret.Generation))))
	})

	It("should create a replicated secret in multiple namespaces", func() {
		ctx := context.Background()
		defer ctx.Done()

		By("create test namespaces")
		ns1 := &corev1.Namespace{}
		ns1.GenerateName = "e2e-"
		Expect(client.Create(ctx, ns1))
		namespaces = append(namespaces, ns1.Name)

		ns2 := &corev1.Namespace{}
		ns2.GenerateName = "e2e-"
		Expect(client.Create(ctx, ns2))
		namespaces = append(namespaces, ns2.Name)

		ns3 := &corev1.Namespace{}
		ns3.GenerateName = "e2e-"
		Expect(client.Create(ctx, ns3))
		namespaces = append(namespaces, ns3.Name)

		secret.Annotations = map[string]string{
			v1alpha1.SecretReplicationNamespacesAnnotation: fmt.Sprintf("%s,%s,%s", ns1.Name, ns2.Name, ns3.Name),
		}
		Expect(client.Update(ctx, secret)).To(Succeed())

		_, err := ctrl.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}})
		Expect(err).ToNot(HaveOccurred())

		newSecret := &corev1.Secret{}
		Expect(client.Get(ctx, types.NamespacedName{Name: secret.Name, Namespace: ns1.Name}, newSecret)).To(Succeed())

		Expect(newSecret.Data).To(Equal(secret.Data))
		Expect(newSecret.Annotations).To(HaveKeyWithValue(v1alpha1.SecretReplicationLasObservedGenerationAnnotation, strconv.Itoa(int(secret.Generation))))

		Expect(client.Get(ctx, types.NamespacedName{Name: secret.Name, Namespace: ns2.Name}, newSecret)).To(Succeed())
		Expect(newSecret.Data).To(Equal(secret.Data))
		Expect(newSecret.Annotations).To(HaveKeyWithValue(v1alpha1.SecretReplicationLasObservedGenerationAnnotation, strconv.Itoa(int(secret.Generation))))

		Expect(client.Get(ctx, types.NamespacedName{Name: secret.Name, Namespace: ns3.Name}, newSecret)).To(Succeed())
		Expect(newSecret.Data).To(Equal(secret.Data))
		Expect(newSecret.Annotations).To(HaveKeyWithValue(v1alpha1.SecretReplicationLasObservedGenerationAnnotation, strconv.Itoa(int(secret.Generation))))
	})

	It("should update an existing secret when data of the source is updated", func() {
		ctx := context.Background()
		defer ctx.Done()

		By("create test namespace")
		ns := &corev1.Namespace{}
		ns.GenerateName = "e2e-"
		Expect(client.Create(ctx, ns))
		namespaces = append(namespaces, ns.Name)

		secret.Annotations = map[string]string{
			v1alpha1.SecretReplicationNamespacesAnnotation: ns.Name,
		}
		Expect(client.Update(ctx, secret)).To(Succeed())

		_, err := ctrl.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}})
		Expect(err).ToNot(HaveOccurred())

		newSecret := &corev1.Secret{}
		Expect(client.Get(ctx, types.NamespacedName{Name: secret.Name, Namespace: ns.Name}, newSecret)).To(Succeed())
		Expect(newSecret.Data).To(Equal(secret.Data))

		secret.Data = nil
		Expect(client.Update(ctx, secret)).To(Succeed())

		_, err = ctrl.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}})
		Expect(err).ToNot(HaveOccurred())

		Expect(client.Get(ctx, types.NamespacedName{Name: secret.Name, Namespace: ns.Name}, newSecret)).To(Succeed())
		Expect(newSecret.Data).To(BeNil())
	})

})
