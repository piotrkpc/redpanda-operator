package result

import (
	"time"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

//// Reconcile successful - don't requeue
//return ctrl.Result{}, nil
//// Reconcile failed due to error - requeue
//return ctrl.Result{}, err
//// Requeue for any reason other than an error
//return ctrl.Result{Requeue: true}, nil

type ReconcileResult interface {
	Completed() bool
	Output() (reconcile.Result, error)
}

type ContinueReconcile struct{}

func (c ContinueReconcile) Completed() bool {
	return false
}
func (c ContinueReconcile) Output() (reconcile.Result, error) {
	panic("there was no Result to return")
}

type Done struct{}

func (d Done) Completed() bool {
	return true
}
func (d Done) Output() (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

type CallBackSoon struct {
	secs int
}

func (c CallBackSoon) Completed() bool {
	return true
}
func (c CallBackSoon) Output() (reconcile.Result, error) {
	t := time.Duration(c.secs) * time.Second
	return reconcile.Result{Requeue: true, RequeueAfter: t}, nil
}

type ErrorOut struct {
	err error
}

func (e ErrorOut) Completed() bool {
	return true
}
func (e ErrorOut) Output() (reconcile.Result, error) {
	return reconcile.Result{}, e.err
}

func Continue() ReconcileResult {
	return ContinueReconcile{}
}

func MakeDone() ReconcileResult {
	return Done{}
}

func RequeueSoon(secs int) ReconcileResult {
	return CallBackSoon{secs: secs}
}

func Error(e error) ReconcileResult {
	return ErrorOut{err: e}
}
