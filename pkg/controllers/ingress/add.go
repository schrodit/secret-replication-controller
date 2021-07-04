package ingressctrl

import (
	"github.com/go-logr/logr"
	"github.com/schrodit/secret-replication-controller/pkg/controllers/errors"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type IngressController struct {
	log    logr.Logger
	client ctrlclient.Client
	scheme *runtime.Scheme
	*errors.ErrorReporter
}

func New(log logr.Logger, client ctrlclient.Client, eventRecorder record.EventRecorder) reconcile.Reconciler {
	return &IngressController{
		log:           log,
		client:        client,
		ErrorReporter: errors.NewErrorReporter(eventRecorder),
	}
}

// AddToMgr adds the secrets reconiler to the given manager
func AddToMgr(log logr.Logger, mgr ctrl.Manager) error {
	c := &IngressController{
		log:           log,
		client:        mgr.GetClient(),
		scheme:        mgr.GetScheme(),
		ErrorReporter: errors.NewErrorReporter(mgr.GetEventRecorderFor("SecretReplicationIngressController")),
	}
	return ctrl.NewControllerManagedBy(mgr).For(&networkingv1.Ingress{}).Complete(c)
}
