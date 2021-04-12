package ingressctrl_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "secret controller test suite")
}

var (
	testenv *envtest.Environment
	client  ctrlclient.Client
)

var _ = BeforeSuite(func() {
	testenv = &envtest.Environment{}

	restConfig, err := testenv.Start()
	Expect(err).ToNot(HaveOccurred())

	client, err = ctrlclient.New(restConfig, ctrlclient.Options{})
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	Expect(testenv.Stop()).To(Succeed())
})
