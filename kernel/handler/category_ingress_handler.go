package handler

import (
	"github.com/kuberator/api/core"
	v1 "github.com/kuberator/api/v1beta1"
	. "github.com/kuberator/kernel/common"
	"github.com/kuberator/kernel/extend"
	. "github.com/kuberator/kernel/util"
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Make make the build-in k8s resource from current component crd
func (component *IngressComponentHandler) Make(source core.CustomResource) (*core.ResourcesLine, error) {
	ingress := &v1beta1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind: string(source.ResourceMeta.GetKind()),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        string(source.ResourceMeta.GetName()),
			Namespace:   source.Crd.GetNamespace(),
			Labels:      Merge(source.ResourceMeta.(*v1.CategoryClusterIngress).Labels, source.Crd.GetLabels()),
			Annotations: Merge(source.ResourceMeta.(*v1.CategoryClusterIngress).Annotations, source.Crd.GetAnnotations()),
			OwnerReferences: []metav1.OwnerReference{
				ToOwnerReference(source)},
		},
		Spec: source.ResourceMeta.(*v1.CategoryClusterIngress).IngressSpec,
	}
	ingress.Labels[CategoryLabel] = string(source.ResourceMeta.GetCategory())

	return &core.ResourcesLine{
		Desired:      ingress,
		ResourceMeta: source.ResourceMeta,
	}, nil
}

// StateFinger convert category state to component state
func (component *IngressComponentHandler) StateFinger(obj client.Object) *v1.ComponentState {
	if obj == nil {
		return v1.NewComponentState(v1.Deleted, "Deleted", map[string]string{})
	}
	is := obj.(*v1beta1.Ingress)
	data := map[string]string{}
	data["spec"] = is.Spec.String()
	return v1.NewComponentState(v1.Success, "ok", data)
}

// OnEvent make and apply will call it
func (component *IngressComponentHandler) OnEvent(event extend.Event) error {
	component.Logger().Info("ingress accept reconcile event", "event", event)
	return nil
}
