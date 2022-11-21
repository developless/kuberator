package kernel

import (
	"github.com/imdario/mergo"
	"github.com/kuberator/api/core"
)

func MergeStage(reconcile *ReconcileContext, resource *core.ResourcesLine) error {
	// make k8s build-in resource
	// get observed resource
	observed, _ := reconcile.GetIfExists(reconcile.Context, reconcile.Namespace, resource.ResourceMeta)
	target, err := reconcile.GetIfExists(reconcile.Context, reconcile.Namespace, resource.ResourceMeta)
	if err != nil {
		reconcile.Log.Error(err, "state finger stage get observer resource cause an error", "category", resource.ResourceMeta.GetCategory(), "name", resource.ResourceMeta.GetName())
	}
	resource.Observed = observed
	if target != nil && resource.Desired != nil {
		err = mergo.Merge(target, resource.Desired, mergo.WithOverride)
		if err != nil {
			reconcile.Log.Error(err, "state finger stage merge resource cause an error", "category", resource.ResourceMeta.GetCategory(), "name", resource.ResourceMeta.GetName())
		}
		resource.Desired = target
	}

	return nil
}
