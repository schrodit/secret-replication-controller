package app

import (
	"flag"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/schrodit/secret-replication-controller/pkg/apis/core/v1alpha1"
	"github.com/schrodit/secret-replication-controller/pkg/logger"
	"github.com/spf13/pflag"
)

type options struct {
	metricsAddr          string
	enableLeaderElection bool
	resyncPeriod         time.Duration
	logConfig            *logger.Config
	alternativePrefixes  []string

	log logr.Logger
}

func (o *options) Complete() error {

	log, err := logger.New(o.logConfig)
	if err != nil {
		return err
	}
	o.log = log

	for _, prefix := range o.alternativePrefixes {
		log.Info(fmt.Sprintf("Configuring alternative 'allNamespaces' annotation %q", v1alpha1.SecretReplicationAllNamespacesAnnotations.Add(prefix)))
		log.Info(fmt.Sprintf("Configuring alternative 'fromNamespace' annotation %q", v1alpha1.SecretReplicationFromNamespaceAnnotations.Add(prefix)))
		log.Info(fmt.Sprintf("Configuring alternative 'namespaces' annotation %q", v1alpha1.SecretReplicationNamespacesAnnotations.Add(prefix)))
	}

	return nil
}

func (o *options) AddFlags(fs *pflag.FlagSet) {
	if fs == nil {
		fs = pflag.CommandLine
	}

	fs.StringVar(&o.metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	fs.BoolVar(&o.enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	fs.DurationVar(&o.resyncPeriod, "resync-period", 10*time.Minute, "Resync interval for the cache if the controller")
	fs.StringArrayVar(&o.alternativePrefixes, "prefix", []string{},
		fmt.Sprintf("define alternate annotation prefixes. Defaults to %q", v1alpha1.DefaultAnnotationPrefix))

	o.logConfig = logger.AddFlags(fs)

	fs.AddGoFlagSet(flag.CommandLine)
}
