package controller

import (
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type secretController struct {
	log    logr.Logger
	client ctrlclient.Client
	scheme *runtime.Scheme
}

func (c *secretController) InjectClient(client ctrlclient.Client) error {
	c.client = client
	return nil
}

func (c *secretController) InjectScheme(s *runtime.Scheme) error {
	c.scheme = s
	return nil
}

// AddToMgr adds the secrets reconiler to the given manager
func AddToMgr(log logr.Logger, mgr ctrl.Manager) error {
	c := &secretController{
		log: log,
	}
	return ctrl.NewControllerManagedBy(mgr).For(&corev1.Secret{}).Complete(c)
}
