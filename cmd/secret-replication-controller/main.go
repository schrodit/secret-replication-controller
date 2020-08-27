package main

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/net/context"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/schrodit/secret-replication-controller/cmd/secret-replication-controller/app"
)

func main() {

	stopCh := ctrl.SetupSignalHandler()
	ctx := &stopChContext{
		parent: context.Background(),
		stopCh: stopCh,
	}

	cmd := app.NewSecretReplicationControllerCmd(ctx)
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type stopChContext struct {
	parent context.Context
	stopCh <-chan struct{}
}

var _ context.Context = &stopChContext{}

func (ctx *stopChContext) Done() <-chan struct{} {
	return ctx.stopCh
}

func (ctx *stopChContext) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (ctx *stopChContext) Err() error {
	return ctx.parent.Err()
}

func (ctx *stopChContext) Value(key interface{}) interface{} {
	return ctx.parent.Value(key)
}
