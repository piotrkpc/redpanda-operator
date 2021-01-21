package pipelines

import (
	"context"
	"fmt"

	"github.com/piotrkpc/redpanda-operator/pkg/reconciliation/result"
)

type PipelineReconciler interface {
	// AddContextTo method should add pipeline specific context to passed ctx
	// All necessary information should be added to the context as it will be use later by Filters
	// that will actually perform reconciliation steps
	AddContextTo(ctx context.Context) context.Context

	// Reconcile method of the pipeline should start the reconciliation
	// trigerring Filter-based processing in proper order
	Reconcile(ctx context.Context) result.ReconcileResult
	UpdateStatus(ctx context.Context) error

	// RegisterFilter adds filter to the pipeline list/slice
	// This method should be called in order that is required by the reconciliation loop
	RegisterFilter(filter Filter) error
}

type Filter interface {
	fmt.Stringer
	// Process performs necessary operations against child resource
	// to implement necessary actions.
	Process(ctx context.Context) result.ReconcileResult
	// UpdateStatus method should update `Status` of the custom resource
	UpdateStatus(ctx context.Context) error
}
