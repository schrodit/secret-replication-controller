package ingressctrl

import (
	"github.com/go-logr/logr"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type IngressController struct {
	log    logr.Logger
	client ctrlclient.Client
	scheme *runtime.Scheme
}

func New(log logr.Logger, client ctrlclient.Client) reconcile.Reconciler {
	return &IngressController{
		log:    log,
		client: client,
	}
}

// AddToMgr adds the secrets reconiler to the given manager
func AddToMgr(log logr.Logger, mgr ctrl.Manager) error {
	c := &IngressController{
		log:    log,
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
	}
	return ctrl.NewControllerManagedBy(mgr).For(&networkingv1.Ingress{}).Complete(c)
}
