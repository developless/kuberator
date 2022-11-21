package kernel

import (
	"github.com/kuberator/api/core"
	"github.com/kuberator/kernel/extend"
	"k8s.io/apimachinery/pkg/types"
)

func PostApplyStage(reconcile *ReconcileContext, command core.ActionCommand, result core.CommandResult) core.CommandResult {
	ch, sh := extend.GetHandler(command.TargetResource.Category)
	// usr define post apply
	if sh != nil {
		return sh.PostApply(core.ComponentArgs{
			Context: reconcile.Context,
			NamespacedName: types.NamespacedName{
				Namespace: reconcile.Namespace,
				Name:      reconcile.Name,
			},
			CustomResource: core.CustomResource{
				ResourceMeta: command.ResourceMeta,
				Crd:          reconcile.Crd,
			},
			Desired: command.TargetResource.Target,
			Logger:  reconcile.Log,
		}, command, result)
	}

	if ch != nil {
		// base post apply
		return ch.PostApply(command, result)
	}

	return result
}
