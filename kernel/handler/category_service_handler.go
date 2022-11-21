package handler

import (
	"fmt"
	"github.com/fatih/structs"
	"github.com/kuberator/api/core"
	v1 "github.com/kuberator/api/v1beta1"
	. "github.com/kuberator/kernel/common"
	"github.com/kuberator/kernel/extend"
	. "github.com/kuberator/kernel/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Make make the build-in k8s resource from current component crd
func (component *ServiceComponentHandler) Make(source core.CustomResource) (*core.ResourcesLine, error) {
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind: string(source.ResourceMeta.GetKind()),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        string(source.ResourceMeta.GetName()),
			Namespace:   source.Crd.GetNamespace(),
			Labels:      Merge(source.ResourceMeta.(*v1.CategoryClusterService).Labels, source.Crd.GetLabels()),
			Annotations: Merge(source.ResourceMeta.(*v1.CategoryClusterService).Annotations, source.Crd.GetAnnotations()),
			OwnerReferences: []metav1.OwnerReference{
				ToOwnerReference(source)},
		},
		Spec: source.ResourceMeta.(*v1.CategoryClusterService).ServiceSpec,
	}
	service.Labels[CategoryLabel] = string(source.ResourceMeta.GetCategory())
	return &core.ResourcesLine{
		Desired:      service,
		ResourceMeta: source.ResourceMeta,
	}, nil
}

// StateFinger convert category state to component state
func (component *ServiceComponentHandler) StateFinger(obj client.Object) *v1.ComponentState {
	if obj == nil {
		return v1.NewComponentState(v1.Deleted, "Deleted", map[string]string{})
	}
	svc := obj.(*corev1.Service)
	spec := structs.Map(&svc.Spec)
	spec["Annotations"] = fmt.Sprintf("%v", svc.Annotations)
	spec["Labels"] = fmt.Sprintf("%v", svc.Labels)
	spec["Ports"] = toString(svc.Spec.Ports...)
	return v1.NewComponentState(v1.Success, "ok", toMap(spec))
}

func toString(ports ...corev1.ServicePort) string {
	target := map[string]string{}
	for _, port := range ports {
		target[port.Name] = fmt.Sprintf("%d", port.Port)
	}
	return v1.ToString(target, "=")
}

func toMap(original map[string]interface{}) map[string]string {
	target := map[string]string{}
	for k, v := range original {
		target[k] = fmt.Sprintf("%+v", v)
	}
	return target
}

// PreApply how to action when apply.
func (component *ServiceComponentHandler) PreApply(observed client.Object, desired client.Object) (*core.ActionCommand, core.CommandResult) {
	act, _ := component.CategoryComponentHandler.PreApply(observed, desired)
	if act.Action == v1.Update {
		if len(observed.(*corev1.Service).Spec.ClusterIP) > 0 {
			desired.(*corev1.Service).Spec.ClusterIP = observed.(*corev1.Service).Spec.ClusterIP
		}
	}
	return act, core.Result()
}

// OnEvent make and apply will call it
func (component *ServiceComponentHandler) OnEvent(event extend.Event) error {
	component.Logger().Info("ingress accept reconcile event", "event", event)
	return nil
}
