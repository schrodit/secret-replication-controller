package app

import (
	"context"
	"fmt"
	"os"

	controller "github.com/schrodit/secret-replication-controller/pkg/controllers/secret"
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
)

// NewSecretReplicationControllerCmd creates a new secret replication controller coommand.
func NewSecretReplicationControllerCmd(ctx context.Context) *cobra.Command {

	opts := &options{}

	cmd := &cobra.Command{
		Use: "secret-replication-controller",

		RunE: func(cmd *cobra.Command, args []string) error {

			if err := opts.Complete(); err != nil {
				return err
			}

			if err := opts.run(ctx); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			return nil
		},
	}

	opts.AddFlags(cmd.Flags())

	return cmd
}

func (o *options) run(ctx context.Context) error {

	restConfig, err := ctrl.GetConfig()
	if err != nil {
		return err
	}

	ctrl.SetLogger(o.log)

	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		MetricsBindAddress: o.metricsAddr,
		Port:               9443,
		LeaderElection:     o.enableLeaderElection,
		LeaderElectionID:   "7d6ea2a1.schrodit.tech",
		SyncPeriod:         &o.resyncPeriod,
		Namespace:          "qa",
	})
	if err != nil {
		return err
	}

	if err := controller.AddToMgr(o.log, mgr); err != nil {
		return err
	}

	return mgr.Start(ctx.Done())
}
