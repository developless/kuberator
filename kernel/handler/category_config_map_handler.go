package handler

import (
	"fmt"
	"github.com/kuberator/api/core"
	v1 "github.com/kuberator/api/v1beta1"
	. "github.com/kuberator/kernel/common"
	"github.com/kuberator/kernel/extend"
	. "github.com/kuberator/kernel/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetConfigMapData(obj core.CategoryComponentObject) map[string]string {
	data := map[string]string{}
	for _, c := range obj.Object.([]*v1.NamedProperties) {
		data[c.Name] = c.Data
	}
	return data
}

// Make make the build-in k8s resource from current component crd
func (component *ConfigMapComponentHandler) Make(source core.CustomResource) (*core.ResourcesLine, error) {
	// Properties which should be provided from real deployed environment.
	meta := source.ResourceMeta.(*core.CategoryComponentObject)
	ref := meta.Reference.(*v1.CategoryClusterComponent)
	if ref.Replicas == nil || *ref.Replicas == 0 {
		return &core.ResourcesLine{
			ResourceMeta: source.ResourceMeta,
		}, nil
	}
	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind: ConfigMap,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: source.Crd.GetNamespace(),
			Name:      string(meta.GetName()),
			OwnerReferences: []metav1.OwnerReference{
				ToOwnerReference(source)},
			Labels:      Merge(nil, source.Crd.GetLabels()),
			Annotations: Merge(nil, source.Crd.GetAnnotations()),
		},
		Data: GetConfigMapData(*meta),
	}

	configMap.Labels[CategoryLabel] = string(meta.GetCategory())
	// reference to the category component.
	configMap.Labels[ReferenceLabel] = string(meta.Reference.GetCategory())
	return &core.ResourcesLine{
		Desired:      configMap,
		ResourceMeta: source.ResourceMeta,
	}, nil
}

// StateFinger convert category state to component state
func (component *ConfigMapComponentHandler) StateFinger(obj client.Object) *v1.ComponentState {
	if obj == nil {
		return v1.NewComponentState(v1.Deleted, "Deleted", map[string]string{})
	}
	cm := obj.(*corev1.ConfigMap)
	return v1.NewComponentState(v1.Success, "ok", cm.Data)
}

// PreApply how to action when apply.
func (component *ConfigMapComponentHandler) PreApply(observed client.Object, desired client.Object) (*core.ActionCommand, core.CommandResult) {
	act, _ := component.CategoryComponentHandler.PreApply(observed, desired)
	if act.Action == v1.Update {
		labels := desired.GetLabels()
		act.Next = GetRestartCommand(desired, labels[ReferenceLabel], 0, fmt.Sprintf("config changed, need restart all the %s pod", labels[ReferenceLabel]))
	}
	return act, core.Result()
}

// OnEvent make and apply will call it
func (component *ConfigMapComponentHandler) OnEvent(event extend.Event) error {
	component.Logger().Info("configMap accept reconcile event", "event", event)
	return nil
}
