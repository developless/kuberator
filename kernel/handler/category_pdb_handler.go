package handler

import (
	"github.com/kuberator/api/core"
	v1 "github.com/kuberator/api/v1beta1"
	. "github.com/kuberator/kernel/common"
	"github.com/kuberator/kernel/extend"
	. "github.com/kuberator/kernel/util"
	"k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func AutoScaleMaxUnavailable(replicas *int32) *intstr.IntOrString {
	var size intstr.IntOrString
	if replicas == nil || *replicas == 1 {
		size = intstr.FromInt(0)
	} else {
		size = intstr.FromInt(int(*replicas - (*replicas/2 + 1)))
	}
	return &size
}

// Make make the build-in k8s resource from current component crd
func (component *PodDisruptionBudgetHandler) Make(source core.CustomResource) (*core.ResourcesLine, error) {
	// Properties which should be provided from real deployed environment.
	meta := source.ResourceMeta.(*core.CategoryComponentObject)
	ref := meta.Reference.(*v1.CategoryClusterComponent)
	if ref == nil || ref.MaxUnavailable == nil {
		return &core.ResourcesLine{
			ResourceMeta: source.ResourceMeta,
		}, nil
	}

	pdb := &v1beta1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:        string(meta.GetName()),
			Namespace:   source.Crd.GetNamespace(),
			Labels:      Merge(nil, source.Crd.GetLabels()),
			Annotations: Merge(nil, source.Crd.GetAnnotations()),
			OwnerReferences: []metav1.OwnerReference{
				ToOwnerReference(source),
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind: PodDisruptionBudget,
		},
		Spec: v1beta1.PodDisruptionBudgetSpec{
			MaxUnavailable: AutoScaleMaxUnavailable(ref.Replicas),
			Selector:       ref.Selector,
		},
	}

	pdb.Labels = Merge(pdb.Labels, GetReferenceLabels(ref, PodDisruptionBudget))

	return &core.ResourcesLine{
		Desired:      pdb,
		ResourceMeta: source.ResourceMeta,
	}, nil
}

// StateFinger convert category state to component state
func (component *PodDisruptionBudgetHandler) StateFinger(obj client.Object) *v1.ComponentState {
	if obj == nil {
		return v1.NewComponentState(v1.Deleted, "Deleted", map[string]string{})
	}
	pdb := obj.(*v1beta1.PodDisruptionBudget)
	data := map[string]string{}
	data["spec"] = pdb.Spec.String()
	return v1.NewComponentState(v1.Success, "ok", data)
}

// OnEvent make and apply will call it.
func (component *PodDisruptionBudgetHandler) OnEvent(event extend.Event) error {
	component.Logger().Info("component accept handler event", "category", event.Category, "name", event.Name, "action", event.Action, "state", event.State)
	return nil
}
