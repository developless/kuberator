package util

import (
	"fmt"
	"github.com/kuberator/api/core"
	v1 "github.com/kuberator/api/v1beta1"
	. "github.com/kuberator/kernel/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func Merge(src1, src2 map[string]string) map[string]string {
	target := make(map[string]string)
	if src1 != nil {
		for k, v := range src1 {
			target[k] = v
		}
	}
	if src2 != nil {
		for k, v := range src2 {
			target[k] = v
		}
	}
	return target
}

func GetComponentName(cluster string, category v1.Category, kind v1.ComponentKind) string {
	return strings.ToLower(fmt.Sprintf("%s-%s-%s", cluster, category, kind))
}

func GetComponentShotName(cluster string, category v1.Category) string {
	return strings.ToLower(fmt.Sprintf("%s-%s", cluster, category))
}

// ToOwnerReference generator the owner reference
func ToOwnerReference(component core.CustomResource) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion:         component.Crd.GetTypeMeta().APIVersion,
		Kind:               component.Crd.GetTypeMeta().Kind,
		Name:               component.Crd.GetName(),
		UID:                component.Crd.GetUID(),
		Controller:         &[]bool{true}[0],
		BlockOwnerDeletion: &[]bool{false}[0],
	}
}

func GetReferenceLabels(ref core.TypedCategoryComponent, kind v1.ComponentKind) map[string]string {
	return map[string]string{
		InstanceLabel:  ref.GetLabels()[InstanceLabel],
		ComponentLabel: ref.GetLabels()[ComponentLabel],
		AppLabel:       ref.GetLabels()[AppLabel],
		CategoryLabel:  strings.ToLower(fmt.Sprintf("%s-%s", ref.GetCategory(), kind)),
		ReferenceLabel: string(ref.GetCategory()),
	}
}

func EnvFinger(envs []corev1.EnvVar) map[string]string {
	target := make(map[string]string)
	for _, e := range envs {
		if len(e.Value) > 0 {
			target[e.Name] = e.Value
		} else if e.ValueFrom != nil {
			target[e.Name] = v1.ToString(EnvResolve(e.ValueFrom), "=")
		}
	}
	return target
}

func EnvResolve(valueFrom *corev1.EnvVarSource) map[string]string {
	target := make(map[string]string)
	if valueFrom.FieldRef != nil {
		target["FieldRef"] = valueFrom.FieldRef.FieldPath
	}
	if valueFrom.ResourceFieldRef != nil {
		target["ResourceFieldRef"] = valueFrom.ResourceFieldRef.String()
	}
	if valueFrom.ConfigMapKeyRef != nil {
		target["ConfigMapKeyRef"] = valueFrom.ConfigMapKeyRef.String()
	}
	if valueFrom.SecretKeyRef != nil {
		target["SecretKeyRef"] = valueFrom.SecretKeyRef.String()
	}
	return target
}

func ContainerFinger(containers ...corev1.Container) map[string]string {
	target := make(map[string]string)
	if containers != nil {
		for i, v := range containers {
			// skip properties
			state := map[string]string{
				"Image":           fmt.Sprintf("%v", v.Image),
				"Command":         fmt.Sprintf("%v", v.Command),
				"Args":            fmt.Sprintf("%v", v.Args),
				"Ports":           fmt.Sprintf("%v", v.Ports),
				"LivenessProbe":   fmt.Sprintf("%v", v.LivenessProbe),
				"ReadinessProbe":  fmt.Sprintf("%v", v.ReadinessProbe),
				"StartupProbe":    fmt.Sprintf("%v", v.StartupProbe),
				"Lifecycle":       fmt.Sprintf("%v", v.Lifecycle),
				"Resources":       fmt.Sprintf("%v", v.Resources),
				"Env":             fmt.Sprintf("%v", v1.ToString(EnvFinger(v.Env), "=")),
				"EnvFrom":         fmt.Sprintf("%v", v.EnvFrom),
				"VolumeMounts":    fmt.Sprintf("%v", v.VolumeMounts),
				"VolumeDevices":   fmt.Sprintf("%v", v.VolumeDevices),
				"SecurityContext": fmt.Sprintf("%v", v.SecurityContext),
			}
			k := v.Name
			if len(k) == 0 {
				k = fmt.Sprintf("container-%d", i)
			}
			target[k] = v1.ToString(state, "=")
		}
	}
	return target
}

func PodSpecFinger(spec corev1.PodSpec) map[string]string {
	target := make(map[string]string)
	target["Volumes"] = fmt.Sprintf("%v", spec.Volumes)
	target["InitContainers"] = fmt.Sprintf("%v", v1.ToString(ContainerFinger(spec.InitContainers...), "="))
	target["Containers"] = fmt.Sprintf("%v", v1.ToString(ContainerFinger(spec.Containers...), "="))
	target["EphemeralContainers"] = fmt.Sprintf("%v", spec.EphemeralContainers)
	target["RestartPolicy"] = fmt.Sprintf("%v", spec.RestartPolicy)
	target["NodeSelector"] = fmt.Sprintf("%v", spec.NodeSelector)
	target["ServiceAccountName"] = fmt.Sprintf("%v", spec.ServiceAccountName)
	target["Affinity"] = fmt.Sprintf("%v", spec.Affinity)
	target["Tolerations"] = fmt.Sprintf("%v", spec.Tolerations)
	target["TopologySpreadConstraints"] = fmt.Sprintf("%v", spec.TopologySpreadConstraints)
	target["SecurityContext"] = fmt.Sprintf("%v", spec.SecurityContext)
	return target
}

func PvcSpecFinger(spec corev1.PersistentVolumeClaimSpec) map[string]string {
	target := make(map[string]string)
	if spec.Selector != nil {
		target["Selector"] = fmt.Sprintf("%v", *spec.Selector)
	}
	target["AccessModes"] = fmt.Sprintf("%v", spec.AccessModes)
	if spec.Resources.Requests.Storage() != nil {
		target["Resources.Request"] = fmt.Sprintf("%v", spec.Resources.Requests.Storage().String())
	}
	if spec.Resources.Limits.Storage() != nil {
		target["Resources.Limit"] = fmt.Sprintf("%v", spec.Resources.Limits.Storage().String())
	}
	if spec.StorageClassName != nil {
		target["StorageClassName"] = fmt.Sprintf("%v", *spec.StorageClassName)
	}
	target["VolumeName"] = fmt.Sprintf("%v", spec.VolumeName)
	if spec.VolumeMode != nil {
		target["VolumeMode"] = fmt.Sprintf("%v", *spec.VolumeMode)
	}
	if spec.DataSource != nil {
		target["DataSource"] = fmt.Sprintf("%v", *spec.DataSource)
	}
	if spec.DataSourceRef != nil {
		target["DataSourceRef"] = fmt.Sprintf("%v", *spec.DataSourceRef)
	}
	return target
}
