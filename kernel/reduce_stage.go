package kernel

import (
	"github.com/kuberator/api/core"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func ReduceStage(reconcile *ReconcileContext, result core.CommandResult) core.CommandResult {
	reconcile.Crd.GetStatus().Gen()
	status := reconcile.Crd.GetStatus()
	var err error
	if err = reconcile.UpdateStatus(reconcile.Context, reconcile.Crd); err != nil {
		if apierrors.IsConflict(err) {
			reconcile.Log.Info("Conflict while updating crd status", "message", err)
			// update reversion and update again.
			err = reconcile.Get(reconcile.Context, reconcile.Crd)
			if err == nil {
				reconcile.Crd.SetStatus(*status)
				err = reconcile.UpdateStatus(reconcile.Context, reconcile.Crd)
				reconcile.Log.Info("retry update crd status", "error", err)
			} else {
				reconcile.Log.Info("get crd then retry update crd status", "error", err)
			}
		} else {
			reconcile.Log.Error(err, "update crd status error", "error", err)
		}
	} else {
		reconcile.Log.Info("update crd status ok.")
	}

	return result.Error(err)
}
