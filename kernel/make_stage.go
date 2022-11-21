package kernel

import (
	"github.com/kuberator/api/core"
	"github.com/kuberator/kernel/extend"
	"k8s.io/apimachinery/pkg/types"
)

func MakeStage(reconcile *ReconcileContext, source core.TypedCategoryComponent) (*core.ResourcesLine, error) {
	// make k8s build-in resource
	ch, sh := extend.GetHandler(source.GetCategory())
	command, err := ch.Make(core.CustomResource{
		ResourceMeta: source,
		Crd:          reconcile.Crd,
	})
	if err != nil {
		reconcile.Log.Error(err, "make build in resource failed", "category", source.GetCategory(), "name", source.GetName())
		return command, err
	}

	// use defined life cycle handler
	if sh != nil {
		command, err = sh.PostMake(core.ComponentArgs{
			Context: reconcile.Context,
			NamespacedName: types.NamespacedName{
				Namespace: reconcile.Namespace,
				Name:      reconcile.Name,
			},
			CustomResource: core.CustomResource{
				Crd:          reconcile.Crd,
				ResourceMeta: source,
			},
			Logger: reconcile.Log,
		}, command)
		if err != nil {
			reconcile.Log.Error(err, "post make build in resource failed", "category", source.GetCategory(), "name", source.GetName())
			return command, err
		}
	}

	return command, err
}
