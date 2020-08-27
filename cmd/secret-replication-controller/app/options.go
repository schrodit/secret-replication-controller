package app

import (
	"flag"
	"time"

	"github.com/go-logr/logr"
	"github.com/schrodit/secret-replication-controller/pkg/logger"
	"github.com/spf13/pflag"
)

type options struct {
	metricsAddr          string
	enableLeaderElection bool
	resyncPeriod         time.Duration
	logConfig            *logger.Config

	log logr.Logger
}

func (o *options) Complete() error {

	log, err := logger.New(o.logConfig)
	if err != nil {
		return err
	}
	o.log = log

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

	o.logConfig = logger.AddFlags(fs)

	fs.AddGoFlagSet(flag.CommandLine)
}
