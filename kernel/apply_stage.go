package kernel

import (
	"github.com/kuberator/api/core"
	v1 "github.com/kuberator/api/v1beta1"
	"github.com/kuberator/kernel/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func Apply(reconcile *ReconcileContext, cmd *core.ActionCommand) core.CommandResult {
	var aerr error
	switch cmd.Action {
	case v1.Create:
		aerr = reconcile.Create(reconcile.Context, cmd.TargetResource.Target)
	case v1.Delete:
		aerr = reconcile.DeleteAllOf(reconcile.Context, cmd.TargetResource)
	case v1.Update:
		aerr = reconcile.Update(reconcile.Context, cmd.TargetResource.Target)
	case v1.Restart:
		aerr = restart(reconcile, cmd)
	case v1.ReCreate:
		aerr = reconcile.ReCreate(reconcile.Context, cmd.TargetResource.Target)
	case v1.FailOver:
		aerr = failOver(reconcile, cmd)
	case v1.RollingUpdate:
	case v1.Recycle:
	case v1.Non:
	}

	return core.Result().Error(aerr)
}

func restart(reconcile *ReconcileContext, cmd *core.ActionCommand) error {
	var pods []corev1.Pod
	if cmd.TargetResource.Extends != nil {
		pods = cmd.TargetResource.Extends.([]corev1.Pod)
	}
	if pods != nil && len(pods) > 0 {
		return reconcile.Restart(reconcile.Context, false, true, pods)
	}

	var podNum int32
	c := reconcile.Crd.GetSpec().GetCategoryResource(cmd.TargetResource.Category)
	if c != nil {
		cc, ok := c.(*v1.CategoryClusterComponent)
		if ok {
			podNum = *cc.Replicas
		}
	}

	name := types.NamespacedName{Namespace: reconcile.Namespace, Name: cmd.TargetResource.Target.GetName()}
	return reconcile.Restart(reconcile.Context, false, true, util.OrderedPod(name, podNum))
}

func failOver(reconcile *ReconcileContext, cmd *core.ActionCommand) error {
	var pods []corev1.Pod
	if cmd.TargetResource.Extends != nil {
		pods = cmd.TargetResource.Extends.([]corev1.Pod)
	}
	if pods != nil && len(pods) > 0 {
		return reconcile.FailOver(reconcile.Context, cmd.TargetResource.Target, pods...)
	}
	return nil
}
