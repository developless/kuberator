package kernel

import (
	"github.com/kuberator/api/core"
	"github.com/kuberator/kernel/extend"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func PreApplyStage(reconcile *ReconcileContext, source core.TypedCategoryComponent, observed, desired client.Object) (*core.ActionCommand, core.CommandResult) {
	ch, sh := extend.GetHandler(source.GetCategory())
	var command *core.ActionCommand
	var result core.CommandResult

	// usr define per apply not return action, it will be use base action.
	if sh != nil {
		command, result = sh.PreApply(core.ComponentArgs{
			Context: reconcile.Context,
			NamespacedName: types.NamespacedName{
				Namespace: reconcile.Namespace,
				Name:      reconcile.Name,
			},
			CustomResource: core.CustomResource{
				Crd:          reconcile.Crd,
				ResourceMeta: source,
			},
			Observed: observed,
			Desired:  desired,
			Logger:   reconcile.Log,
		})

		if result.NotEmpty() {
			return command, result
		}
	}

	if command == nil && ch != nil {
		return ch.PreApply(observed, desired)
	}

	return command, result
}
