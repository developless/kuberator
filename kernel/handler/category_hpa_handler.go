package handler

import (
	"github.com/kuberator/api/core"
	v1 "github.com/kuberator/api/v1beta1"
	. "github.com/kuberator/kernel/common"
	"github.com/kuberator/kernel/extend"
	. "github.com/kuberator/kernel/util"
	"k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Make make the build-in k8s resource from current component crd
func (component *HorizontalPodAutoscalerHandler) Make(source core.CustomResource) (*core.ResourcesLine, error) {
	// Properties which should be provided from real deployed environment.
	meta := source.ResourceMeta.(*core.CategoryComponentObject)
	ref := meta.Reference.(*v1.CategoryClusterComponent)
	if ref == nil || ref.Replicas == nil || *ref.Replicas == 0 || ref.MaxReplicas == nil || *ref.MaxReplicas == 0 {
		return &core.ResourcesLine{
			ResourceMeta: source.ResourceMeta,
		}, nil
	}

	hpa := &v2beta2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: source.Crd.GetNamespace(),
			Name:      string(meta.GetName()),
			OwnerReferences: []metav1.OwnerReference{
				ToOwnerReference(source)},
			Labels:      Merge(nil, source.Crd.GetLabels()),
			Annotations: Merge(nil, source.Crd.GetAnnotations()),
		},
		TypeMeta: metav1.TypeMeta{
			Kind: HorizontalPodAutoscaler,
		},
		Spec: v2beta2.HorizontalPodAutoscalerSpec{
			MaxReplicas: *ref.MaxReplicas,
			MinReplicas: ref.Replicas,
			Metrics:     ref.Metrics,
			Behavior:    ref.Behavior,
		},
	}

	hpa.Labels = Merge(hpa.Labels, GetReferenceLabels(ref, HorizontalPodAutoscaler))
	return &core.ResourcesLine{
		Desired:      hpa,
		ResourceMeta: source.ResourceMeta,
	}, nil
}

// StateFinger convert category state to component state
func (component *HorizontalPodAutoscalerHandler) StateFinger(obj client.Object) *v1.ComponentState {
	if obj == nil {
		return v1.NewComponentState(v1.Deleted, "Deleted", map[string]string{})
	}
	hpa := obj.(*v2beta2.HorizontalPodAutoscaler)
	data := map[string]string{}
	data["spec"] = hpa.Spec.String()
	return v1.NewComponentState(v1.Success, "ok", data)
}

// OnEvent make and apply will call it
func (component *HorizontalPodAutoscalerHandler) OnEvent(event extend.Event) error {
	component.Logger().Info("configMap accept reconcile event", "event", event)
	return nil
}
