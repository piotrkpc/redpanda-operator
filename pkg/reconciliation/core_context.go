package reconciliation

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var _ context.Context = &CoreReconciliationContext{}

type CoreReconciliationContext struct {
	Request   *ctrl.Request
	Client    client.Client
	Scheme    *runtime.Scheme
	ReqLogger logr.Logger
	Ctx       context.Context
}

func (c *CoreReconciliationContext) Deadline() (deadline time.Time, ok bool) {
	return c.Ctx.Deadline()
}

func (c *CoreReconciliationContext) Done() <-chan struct{} {
	return c.Ctx.Done()
}

func (c *CoreReconciliationContext) Err() error {
	return c.Ctx.Err()
}

func (c *CoreReconciliationContext) Value(key interface{}) interface{} {
	return c.Ctx.Value(key)
}
