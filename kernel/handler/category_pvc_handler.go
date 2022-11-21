package handler

import (
	"context"
	"fmt"
	"github.com/kuberator/api/core"
	v1 "github.com/kuberator/api/v1beta1"
	. "github.com/kuberator/kernel/common"
	"github.com/kuberator/kernel/extend"
	. "github.com/kuberator/kernel/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
)

// Make make the build-in k8s resource from current component crd
func (component *PersistentVolumeClaimHandler) Make(source core.CustomResource) (*core.ResourcesLine, error) {
	// Properties which should be provided from real deployed environment.
	meta := source.ResourceMeta.(*core.CategoryComponentObject)
	ref := meta.Reference.(*v1.CategoryClusterComponent)
	if ref == nil || ref.Replicas == nil || *ref.Replicas == 0 {
		return &core.ResourcesLine{
			ResourceMeta: source.ResourceMeta,
		}, nil
	}

	var resources *core.ResourcesLine
	for i := int32(0); i < *ref.Replicas; i++ {
		name := fmt.Sprintf("pvc-%s-%s-%d", ref.Labels[InstanceLabel], ref.GetCategory(), i)
		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: source.Crd.GetNamespace(),
				Name:      name,
				OwnerReferences: []metav1.OwnerReference{
					ToOwnerReference(source)},
				Labels:      Merge(nil, source.Crd.GetLabels()),
				Annotations: Merge(nil, source.Crd.GetAnnotations()),
			},
			TypeMeta: metav1.TypeMeta{
				Kind: PersistentVolumeClaim,
			},
			Spec: *ref.PersistentVolumeClaim,
		}

		pvc.Labels = Merge(pvc.Labels, GetReferenceLabels(ref, PersistentVolumeClaim))

		resourceMeta := source.ResourceMeta
		resourceMeta.SetName(v1.ComponentName(name))
		rs := &core.ResourcesLine{
			Desired:      pvc,
			ResourceMeta: resourceMeta,
		}
		if resources == nil {
			resources = rs
		} else {
			resources.Append(rs)
		}
	}

	return resources, nil
}

// StateFinger convert category state to component state
func (component *PersistentVolumeClaimHandler) StateFinger(obj client.Object) *v1.ComponentState {
	if obj == nil {
		return v1.NewComponentState(v1.Deleted, "Deleted", map[string]string{})
	}
	data := PvcSpecFinger(obj.(*corev1.PersistentVolumeClaim).Spec)
	return v1.NewComponentState(v1.Success, "ok", data)
}

func (component *PersistentVolumeClaimHandler) PersistentVolumeClaimHorizontalScale(observed *corev1.PersistentVolumeClaim, desired *corev1.PersistentVolumeClaim) *core.ActionCommand {
	if observed == nil || desired == nil {
		return nil
	}
	// 1. update pvc cr
	actions := &core.ActionCommand{
		Action: v1.Recycle,
		TargetResource: &core.ReferenceObject{
			Category: v1.Category(desired.Labels[CategoryLabel]),
			Target:   desired,
		},
		Callback: func(result *core.CommandResult, cli client.Client, i ...interface{}) error {
			pvcs, er := PvcList(cli, desired)
			if len(pvcs) == 0 {
				component.Logger().Info("not found pvc items, may be the select labels not right", "labels", desired.GetLabels())
				return nil
			}

			sts := &appsv1.StatefulSet{
				TypeMeta: metav1.TypeMeta{
					Kind: StatefulSet,
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: desired.GetNamespace(),
					Name:      fmt.Sprintf("%s-%s", desired.GetLabels()[InstanceLabel], desired.GetLabels()[ReferenceLabel]),
				},
			}
			er = cli.Get(context.Background(), client.ObjectKeyFromObject(sts), sts)
			if er != nil && !(apierrors.IsNotFound(er) || apierrors.IsGone(er)) {
				return er
			}

			podNum := sts.Spec.Replicas
			if podNum == nil {
				zero := int32(0)
				podNum = &zero
			}

			sort.Slice(pvcs, func(i, j int) bool {
				return pvcs[i].Name > pvcs[j].Name
			})
			// down scale
			if len(pvcs) > int(*podNum) {
				index := len(pvcs) - int(*podNum)
				pvcs = pvcs[:index]
				for _, p := range pvcs {
					er = cli.Delete(context.Background(), &p)
					if er != nil {
						return er
					}
				}
			}
			return nil
		},
	}
	return actions
}

// Visitation when the state not change, visitation the relationship resource.
func (component *PersistentVolumeClaimHandler) Visitation(args core.ComponentArgs) *core.ActionCommand {
	return component.PersistentVolumeClaimHorizontalScale(args.Observed.(*corev1.PersistentVolumeClaim), args.Desired.(*corev1.PersistentVolumeClaim))
}

// PreApply how to action when apply.
func (component *PersistentVolumeClaimHandler) PreApply(observed client.Object, desired client.Object) (*core.ActionCommand, core.CommandResult) {
	//TODO: nothing todo, may pvc scale in here.
	return nil, core.Result()
}

// OnEvent make and apply will call it
func (component *PersistentVolumeClaimHandler) OnEvent(event extend.Event) error {
	component.Logger().Info("configMap accept reconcile event", "event", event)
	return nil
}
