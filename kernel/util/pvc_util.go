package util

import (
	"context"
	"errors"
	"github.com/kuberator/api/core"
	v1 "github.com/kuberator/api/v1beta1"
	. "github.com/kuberator/kernel/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func MakePersistentVolumeClaim(sts appsv1.StatefulSet, source *v1.CategoryClusterComponent) *corev1.PersistentVolumeClaim {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "pvc",
			Namespace:       sts.Namespace,
			Labels:          Merge(nil, GetReferenceLabels(source, PersistentVolumeClaim)),
			OwnerReferences: sts.OwnerReferences,
		},
		TypeMeta: metav1.TypeMeta{Kind: PersistentVolumeClaim},
		Spec:     *source.PersistentVolumeClaim,
	}
	if pvc.Spec.VolumeMode == nil {
		mode := corev1.PersistentVolumeFilesystem
		pvc.Spec.VolumeMode = &mode
	}
	return pvc
}

func ValidatePersistentVolumeClaim(observed *corev1.PersistentVolumeClaim, desired *corev1.PersistentVolumeClaim) (int, error) {
	// observed is nil
	if observed == nil {
		if desired == nil {
			return 0, nil
		}
		return 1, nil
	}
	if desired == nil {
		return -1, errors.New("pvc not support down scale")
	}

	// observed request is nil
	if observed.Spec.Resources.Requests == nil || observed.Spec.Resources.Requests.Storage() == nil {
		if desired.Spec.Resources.Requests == nil || desired.Spec.Resources.Requests.Storage() == nil {
			return 0, nil
		}
		return 1, nil
	}

	// desired request is nil
	if desired.Spec.Resources.Requests == nil || desired.Spec.Resources.Requests.Storage() == nil {
		return -1, errors.New("request storage must be setting")
	}

	request := desired.Spec.Resources.Requests.Storage().Cmp(*observed.Spec.Resources.Requests.Storage())
	if request == -1 {
		return request, errors.New("pvc not support down scale")
	}

	// limit is not be modify, so the new size request must be less than the old limit.
	if observed.Spec.Resources.Limits != nil && observed.Spec.Resources.Limits.Storage() != nil {
		limit := desired.Spec.Resources.Requests.Storage().Cmp(*observed.Spec.Resources.Limits.Storage())
		if limit == 1 {
			return request, errors.New("pvc up scale request more than current limit")
		}
	}

	return request, nil
}

func PvcList(cli client.Client, template *corev1.PersistentVolumeClaim) ([]corev1.PersistentVolumeClaim, error) {
	var pvcs corev1.PersistentVolumeClaimList
	namespace := client.InNamespace(template.GetNamespace())
	matchLabels := client.MatchingLabels(map[string]string{
		InstanceLabel:  template.GetLabels()[InstanceLabel],
		ComponentLabel: template.GetLabels()[ComponentLabel],
		AppLabel:       template.GetLabels()[AppLabel],
		ReferenceLabel: template.GetLabels()[ReferenceLabel],
	})
	err := cli.List(context.Background(), &pvcs, matchLabels, namespace)
	if err != nil {
		return nil, err
	}
	return pvcs.Items, nil
}

func PersistentVolumeClaimVectorScale(observed *corev1.PersistentVolumeClaim, desired *corev1.PersistentVolumeClaim) (*core.ActionCommand, error) {
	scale, err := ValidatePersistentVolumeClaim(observed, desired)
	if err != nil || scale == 0 {
		return nil, err
	}

	// 1. update pvc cr
	actions := &core.ActionCommand{
		Action: v1.RollingUpdate,
		TargetResource: &core.ReferenceObject{
			Category: v1.Category(desired.Labels[CategoryLabel]),
			Target:   desired,
		},
		Validate: func(cli client.Client, i ...interface{}) error {
			var sc storagev1.StorageClass
			scn := desired.Spec.StorageClassName
			if er := cli.Get(context.Background(), types.NamespacedName{Name: *scn}, &sc); er != nil {
				return er
			}
			if sc.AllowVolumeExpansion == nil || !*sc.AllowVolumeExpansion {
				return errors.New("StorageClass not support Expansion")
			}
			return nil
		},
		Callback: func(result *core.CommandResult, cli client.Client, i ...interface{}) error {
			pvcs, er := PvcList(cli, desired)
			if len(pvcs) == 0 {
				return nil
			}
			for _, pvc := range pvcs {
				pvc.Spec.Resources.Requests[corev1.ResourceStorage] = desired.Spec.Resources.Requests[corev1.ResourceStorage]
				er = cli.Update(context.Background(), &pvc)
				if er != nil {
					return er
				}
			}
			return nil
		},
	}

	return actions, nil
}
