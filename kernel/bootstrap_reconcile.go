package kernel

import (
	"context"
	"github.com/kuberator/api/core"
	v1 "github.com/kuberator/api/v1beta1"
	. "github.com/kuberator/kernel/common"
	"github.com/kuberator/kernel/util"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"runtime/debug"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// ReconcileContext component crd client
type ReconcileContext struct {
	Scheme *runtime.Scheme `json:"scheme,omitempty"`
	ctrl.Request
	Context  context.Context
	Recorder record.EventRecorder
	Crd      core.BasicCrd
	util.ReconcileClient
}

// Reconcile bootstrap reconcile access
func Reconcile(reconcile *ReconcileContext) core.CommandResult {

	defer func() {
		if e := recover(); e != nil {
			debug.PrintStack()
			reconcile.Log.Error(e.(error), "runtime panic!")
		}
	}()

	// get crd
	err := reconcile.Get(reconcile.Context, reconcile.Crd)
	if err != nil {
		// crd is deleted.
		if client.IgnoreNotFound(err) == nil {
			reconcile.Log.Info("crd not found, may be it is deleted!")
			return core.Result().WithDelete()
		}
		reconcile.Log.Error(err, "crd get error or not found, please check and install it first!")
		return core.Result().Error(err)
	}

	if reconcile.Crd.GetLabels()[InstancePauseLabel] == "true" {
		reconcile.Log.Info("instance reconcile status is pause, requeue the event after 60s")
		return core.Result().WithRequeueAfter(60 * time.Second)
	}

	delete(reconcile.Crd.GetAnnotations(), LastAppliedAnnotation)
	if reconcile.Crd.GetStatus().ComponentStatus == nil {
		reconcile.Crd.SetStatus(*v1.NewClusterComponentStatus())
	}

	if !reconcile.Crd.GetDeletionTimestamp().IsZero() {
		reconcile.Log.Info("resource marked delete", "deletionTimestamp", reconcile.Crd.GetDeletionTimestamp())
		return core.Result()
	}

	reconcile.Log.Info("crd get ok begin construct pipeline...", "crd", reconcile.Crd)

	// pipeline construct and action.
	result := Compile(reconcile).
		WithMakeFunc(func(reconcile *ReconcileContext, task core.TypedCategoryComponent) (*core.ResourcesLine, error) {
			return MakeStage(reconcile, task)
		}).
		WithMergeFunc(func(reconcile *ReconcileContext, resource *core.ResourcesLine) error {
			return MergeStage(reconcile, resource)
		}).
		WithStateFingerFunc(func(reconcile *ReconcileContext, task core.TypedCategoryComponent, observed, desired client.Object) (bool, *v1.ComponentState) {
			return StateFingerStage(reconcile, task, observed, desired)
		}).
		WithVisitationFunc(func(reconcile *ReconcileContext, task core.TypedCategoryComponent, observed, desired client.Object) *core.ActionCommand {
			return VisitationStage(reconcile, task, observed, desired)
		}).
		WithPreApplyFunc(func(reconcile *ReconcileContext, task core.TypedCategoryComponent, observed, desired client.Object) (*core.ActionCommand, core.CommandResult) {
			return PreApplyStage(reconcile, task, observed, desired)
		}).
		WithApplyFunc(func(reconcile *ReconcileContext, command *core.ActionCommand) core.CommandResult {
			return Apply(reconcile, command)
		}).
		WithPostApplyFunc(func(reconcile *ReconcileContext, command core.ActionCommand, result core.CommandResult) core.CommandResult {
			return PostApplyStage(reconcile, command, result)
		}).
		WithReduceFunc(func(reconcile *ReconcileContext, result core.CommandResult) core.CommandResult {
			return ReduceStage(reconcile, result)
		}).
		Compute()

	result.Print(reconcile.Log)

	return result
}
