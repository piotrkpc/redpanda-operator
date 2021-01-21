package k8s

import (
	"context"
	"github.com/piotrkpc/redpanda-operator/pkg/reconciliation/pipelines"
	"github.com/piotrkpc/redpanda-operator/pkg/reconciliation/result"
)

var _ pipelines.PipelineReconciler = &K8sPipeline{}

type K8sPipeline struct {
	Filters []pipelines.Filter
}

func (k *K8sPipeline) AddContextTo(ctx context.Context) context.Context {
	panic("implement me")
}

func (k *K8sPipeline) Reconcile(ctx context.Context) result.ReconcileResult {
	for _, filter := range k.Filters {
		reconcileResult := filter.Process(ctx)
		if reconcileResult.Completed() {
			return reconcileResult
		}
		err := filter.UpdateStatus(ctx)
		if err != nil {
		}
	}
	return result.Continue()
}

func (k *K8sPipeline) UpdateStatus(ctx context.Context) error {
	panic("implement me")
}

func (k *K8sPipeline) RegisterFilter(filter pipelines.Filter) error {
	panic("implement me")
}
