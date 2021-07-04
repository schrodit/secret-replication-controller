package secretctrl

import (
	"github.com/go-logr/logr"
	"github.com/schrodit/secret-replication-controller/pkg/controllers/errors"
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
	*errors.ErrorReporter
}

// AddToMgr adds the secrets reconiler to the given manager
func AddToMgr(log logr.Logger, mgr manager.Manager) error {
	c := &secretController{
		log:           log,
		client:        mgr.GetClient(),
		scheme:        mgr.GetScheme(),
		ErrorReporter: errors.NewErrorReporter(mgr.GetEventRecorderFor("SecretReplicationSecretController")),
	}
	return ctrl.NewControllerManagedBy(mgr).For(&corev1.Secret{}).Complete(c)
}
