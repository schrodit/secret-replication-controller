package secretctrl

import (
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type secretController struct {
	log    logr.Logger
	client ctrlclient.Client
	scheme *runtime.Scheme
}

// AddToMgr adds the secrets reconiler to the given manager
func AddToMgr(log logr.Logger, mgr manager.Manager) error {
	c := &secretController{
		log:    log,
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
	}
	return ctrl.NewControllerManagedBy(mgr).For(&corev1.Secret{}).Complete(c)
}
