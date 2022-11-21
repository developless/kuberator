package kernel

import (
	"github.com/kuberator/api/core"
	"github.com/kuberator/kernel/extend"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func VisitationStage(reconcile *ReconcileContext, source core.TypedCategoryComponent, observed, desired client.Object) *core.ActionCommand {
	ch, sh := extend.GetHandler(source.GetCategory())
	var command *core.ActionCommand

	args := core.ComponentArgs{
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
	}

	// usr define per apply not return action, it will be use base action.
	if sh != nil {
		command = sh.Visitation(args)
	}

	if command == nil && ch != nil {
		return ch.Visitation(args)
	}

	return command
}
