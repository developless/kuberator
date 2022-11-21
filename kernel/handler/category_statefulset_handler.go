package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/kuberator/api/core"
	v1 "github.com/kuberator/api/v1beta1"
	. "github.com/kuberator/kernel/common"
	"github.com/kuberator/kernel/extend"
	. "github.com/kuberator/kernel/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
)

func getCrd(source core.CustomResource) *v1.CategoryClusterComponent {
	return source.ResourceMeta.(*v1.CategoryClusterComponent)
}

// MergeConf the desired configMap.
func (component *StatefulSetClusterComponent) MergeConf(source core.CustomResource) []v1.NamedProperties {
	// Properties which should be provided from real deployed environment.
	conf := map[string]*v1.NamedProperties{}
	if source.Crd.GetSpec().Conf != nil {
		for _, c := range source.Crd.GetSpec().Conf {
			conf[c.PropertiesName()] = c
		}
	}

	// pod conf
	if getCrd(source).Properties != nil {
		for _, v := range getCrd(source).Properties {
			if conf[v.PropertiesName()] == nil {
				conf[v.PropertiesName()] = v
			}
		}
	}

	if len(conf) == 0 {
		return nil
	}

	keys := make([]string, len(conf))
	i := 0
	for k := range conf {
		keys[i] = k
		i = i + 1
	}
	sort.Strings(keys)

	var confList []v1.NamedProperties
	for _, key := range keys {
		confList = append(confList, *conf[key])
	}

	return confList
}

// Make make the build-in k8s resource from current component crd
func (component *StatefulSetClusterComponent) Make(source core.CustomResource) (*core.ResourcesLine, error) {
	replicas := getCrd(source).Replicas
	if replicas == nil || *replicas == 0 {
		return &core.ResourcesLine{ResourceMeta: source.ResourceMeta}, nil
	}
	// build-in statefulSet
	statefulSet := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind: string(source.ResourceMeta.GetKind()),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        string(source.ResourceMeta.GetName()),
			Namespace:   source.Crd.GetNamespace(),
			Labels:      Merge(source.Crd.GetLabels(), getCrd(source).Labels),
			Annotations: Merge(source.Crd.GetAnnotations(), getCrd(source).Annotations),
			OwnerReferences: []metav1.OwnerReference{
				ToOwnerReference(source)},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      Merge(getCrd(source).Labels, getCrd(source).Template.Labels),
					Annotations: Merge(getCrd(source).Annotations, getCrd(source).Template.Annotations),
				},
				Spec: getCrd(source).Template.Spec,
			},
			Selector:            getCrd(source).Selector,
			UpdateStrategy:      getCrd(source).UpdateStrategy,
			PodManagementPolicy: getCrd(source).PodManagementPolicy,
			ServiceName:         GetComponentShotName(source.Crd.GetName(), v1.Category(getCrd(source).ServiceName)),
		},
	}
	var RevisionHistoryLimit = int32(10)
	if len(statefulSet.Spec.UpdateStrategy.Type) == 0 {
		//OnDelete 更新策略实现了传统（1.7 之前）行为，它也是默认的更新策略。 当你选择这个更新策略并修改 StatefulSet 的 .spec.template 字段时，StatefulSet 控制器将不会自动更新 Pod。
		statefulSet.Spec.UpdateStrategy.Type = appsv1.OnDeleteStatefulSetStrategyType
	}
	if len(statefulSet.Spec.PodManagementPolicy) == 0 {
		statefulSet.Spec.PodManagementPolicy = appsv1.ParallelPodManagement
	}
	if statefulSet.Spec.RevisionHistoryLimit == nil {
		statefulSet.Spec.RevisionHistoryLimit = &RevisionHistoryLimit
	}

	//com label
	statefulSet.Labels[InstanceLabel] = source.Crd.GetName()
	statefulSet.Labels[CategoryLabel] = string(getCrd(source).GetCategory())
	statefulSet.Spec.Template.Labels[InstanceLabel] = source.Crd.GetName()
	statefulSet.Spec.Template.Labels[CategoryLabel] = string(getCrd(source).GetCategory())

	confList := component.MergeConf(source)

	// default env
	envs := []corev1.EnvVar{
		{
			Name: Namespace,
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
		{
			Name: NodeName,
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "spec.nodeName",
				},
			},
		},
		{
			Name: HostIp,
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.hostIP",
				},
			},
		},
		{
			Name:  AppName,
			Value: source.Crd.GetName(),
		},
		{
			Name:  Category,
			Value: string(source.ResourceMeta.GetCategory()),
		},
		{
			Name:  ClusterDomain,
			Value: os.Getenv(ClusterDomain),
		},
		{
			Name:  PeerService,
			Value: fmt.Sprintf("%s.%s.svc.%s", statefulSet.Spec.ServiceName, source.Crd.GetNamespace(), os.Getenv(ClusterDomain)),
		},
	}

	// pvc
	pvc := MakePersistentVolumeClaim(*statefulSet, getCrd(source))
	if pvc != nil {
		statefulSet.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
			*pvc,
		}
	}

	// init containers
	statefulSet.Spec.Template.Spec.InitContainers = mergeDefaultEnv(envs, statefulSet.Spec.Template.Spec.InitContainers)

	// container
	statefulSet.Spec.Template.Spec.Containers = mergeDefaultEnv(envs, statefulSet.Spec.Template.Spec.Containers)
	for c := range statefulSet.Spec.Template.Spec.Containers {
		var spec = &statefulSet.Spec.Template.Spec
		var container = &spec.Containers[c]

		// volume
		vl, vm := BuildConfigMapMount(getCrd(source).Category, source.Crd.GetName(), confList)
		if vl != nil {
			spec.Volumes = append(spec.Volumes, vl...)
		}

		// amount
		if vm != nil {
			container.VolumeMounts = append(container.VolumeMounts, vm...)
		}
	}

	return &core.ResourcesLine{
		Desired:      statefulSet,
		ResourceMeta: source.ResourceMeta,
	}, nil
}

func mergeDefaultEnv(env []corev1.EnvVar, containers []corev1.Container) []corev1.Container {
	for i, c := range containers {
		var target []corev1.EnvVar
		target = append(target, env...)
		target = append(target, c.Env...)
		containers[i].Env = target
	}
	return containers
}

// StateFinger the crd state
func (component *StatefulSetClusterComponent) StateFinger(obj client.Object) *v1.ComponentState {
	data := map[string]string{}

	if obj == nil {
		return v1.NewComponentState(v1.Deleted, "Deleted", data)
	}
	ss := obj.(*appsv1.StatefulSet)

	data["Replicas"] = fmt.Sprintf("%v", *ss.Spec.Replicas)
	if ss.Spec.Selector != nil {
		data["Selector"] = fmt.Sprintf("%v", *ss.Spec.Selector)
	}
	data["Annotations"] = fmt.Sprintf("%v", ss.Spec.Template.Annotations)
	data["Labels"] = fmt.Sprintf("%v", ss.Spec.Template.Labels)
	data["Template"] = fmt.Sprintf("%v", PodSpecFinger(ss.Spec.Template.Spec))
	data["ServiceName"] = fmt.Sprintf("%v", ss.Spec.ServiceName)
	data["PodManagementPolicy"] = fmt.Sprintf("%v", ss.Spec.PodManagementPolicy)
	if ss.Spec.RevisionHistoryLimit != nil {
		data["RevisionHistoryLimit"] = fmt.Sprintf("%v", *ss.Spec.RevisionHistoryLimit)
	}
	data["MinReadySeconds"] = fmt.Sprintf("%v", ss.Spec.MinReadySeconds)

	if ss.Spec.VolumeClaimTemplates != nil || len(ss.Spec.VolumeClaimTemplates) > 0 {
		for i, c := range ss.Spec.VolumeClaimTemplates {
			data[fmt.Sprintf("VolumeClaimTemplates[%d]", i)] = v1.ToString(PvcSpecFinger(c.Spec), "=")
		}
	}

	return v1.NewComponentState(v1.Success, "ok", data)
}

func (component *StatefulSetClusterComponent) RestartCheck(observed client.Object, desired client.Object) *core.ActionCommand {
	if observed != nil && desired != nil {
		metaO := v1.ToString(PodSpecFinger(observed.(*appsv1.StatefulSet).Spec.Template.Spec), "=")
		metaD := v1.ToString(PodSpecFinger(desired.(*appsv1.StatefulSet).Spec.Template.Spec), "=")
		if metaO != metaD {
			component.Logger().Info("StatefulSet pod template is changed")
			PrintFingerDiff(metaO, metaD)
			labels := desired.GetLabels()

			replicas := *observed.(*appsv1.StatefulSet).Spec.Replicas
			if replicas > *desired.(*appsv1.StatefulSet).Spec.Replicas {
				replicas = *desired.(*appsv1.StatefulSet).Spec.Replicas
			}

			return GetRestartCommand(desired, labels[CategoryLabel], replicas, fmt.Sprintf("StatefulSet pod template is changed, need restart all the %s pod", labels[CategoryLabel]))
		}
	}
	return nil
}

func (component *StatefulSetClusterComponent) RecreateCheck(observed client.Object, desired client.Object) *core.ActionCommand {
	if observed != nil && desired != nil {
		var recreate *core.ActionCommand
		dataObserved := map[string]string{}
		dataDesired := map[string]string{}
		if observed.(*appsv1.StatefulSet).Spec.Selector != nil {
			dataObserved["Selector"] = fmt.Sprintf("%v", *observed.(*appsv1.StatefulSet).Spec.Selector)
		}
		if desired.(*appsv1.StatefulSet).Spec.Selector != nil {
			dataDesired["Selector"] = fmt.Sprintf("%v", *desired.(*appsv1.StatefulSet).Spec.Selector)
		}
		if observed.(*appsv1.StatefulSet).Spec.VolumeClaimTemplates != nil || len(observed.(*appsv1.StatefulSet).Spec.VolumeClaimTemplates) > 0 {
			for i, c := range observed.(*appsv1.StatefulSet).Spec.VolumeClaimTemplates {
				dataObserved[fmt.Sprintf("VolumeClaimTemplates[%d]", i)] = fmt.Sprintf("%v", c.Spec.String())
			}
		}
		if desired.(*appsv1.StatefulSet).Spec.VolumeClaimTemplates != nil || len(desired.(*appsv1.StatefulSet).Spec.VolumeClaimTemplates) > 0 {
			for i, c := range desired.(*appsv1.StatefulSet).Spec.VolumeClaimTemplates {
				dataObserved[fmt.Sprintf("VolumeClaimTemplates[%d]", i)] = fmt.Sprintf("%v", c.Spec.String())
			}
		}

		metaObserved := v1.ToString(dataObserved, "=")
		metaDesired := v1.ToString(dataDesired, "=")
		if metaObserved != metaDesired {
			message := "StatefulSet selector is changed or pvc scale, need recreate it"
			component.Logger().Info(message)
			PrintFingerDiff(metaObserved, metaDesired)

			act, err := PersistentVolumeClaimVectorScale(&observed.(*appsv1.StatefulSet).Spec.VolumeClaimTemplates[0], &desired.(*appsv1.StatefulSet).Spec.VolumeClaimTemplates[0])
			if err != nil {
				component.Logger().Error(err, "pvc checked error")
				return nil
			}

			replicas := *observed.(*appsv1.StatefulSet).Spec.Replicas
			size := *desired.(*appsv1.StatefulSet).Spec.Replicas

			effective := replicas
			step := replicas
			// downscale
			if replicas > size {
				step = replicas - 1
				effective = size
			} else if replicas < size {
				// upscale
				step = replicas + 1
			}
			desired.(*appsv1.StatefulSet).Spec.Replicas = &step

			recreate = &core.ActionCommand{
				Action:  v1.ReCreate,
				Message: message,
				TargetResource: &core.ReferenceObject{
					Category: v1.Category(desired.GetLabels()[CategoryLabel]),
					Target:   desired,
				},
			}

			if act != nil {
				return act.Append(recreate).Append(GetRestartCommand(desired, desired.GetLabels()[CategoryLabel], effective, message))
			}

			return recreate
		}
	}
	return nil
}

func (component *StatefulSetClusterComponent) OrderedReadyPodManagement(observed client.Object, desired client.Object) *core.ActionCommand {
	replicas := *observed.(*appsv1.StatefulSet).Spec.Replicas
	size := *desired.(*appsv1.StatefulSet).Spec.Replicas
	if replicas == 0 || size == 0 || replicas == size {
		return nil
	}

	count := replicas
	// downscale
	if replicas > size {
		count = replicas - 1
	} else if replicas < size {
		// upscale
		count = replicas + 1
	}

	desired.(*appsv1.StatefulSet).Spec.Replicas = &count

	return &core.ActionCommand{
		Action: v1.Update,
		Validate: func(cli client.Client, i ...interface{}) error {
			var pods corev1.PodList
			namespace := client.InNamespace(desired.GetNamespace())
			matchLabels := client.MatchingLabels(desired.(*appsv1.StatefulSet).Spec.Selector.MatchLabels)
			err := client.IgnoreNotFound(cli.List(context.Background(), &pods, matchLabels, namespace))
			if err != nil {
				return err
			}
			if len(pods.Items) != 0 && !IsPodReady(pods.Items...) {
				return errors.New("not all the pods are running state")
			}
			return nil
		},
		TargetResource: &core.ReferenceObject{
			Target: desired,
		},
	}
}

// Visitation when the state not change, visitation the relationship resource.
func (component *StatefulSetClusterComponent) Visitation(args core.ComponentArgs) *core.ActionCommand {
	if args.Observed == nil || args.Desired == nil {
		return nil
	}
	return GetFailOverCommand(args.Desired)
}

// PreApply how to action when apply.
func (component *StatefulSetClusterComponent) PreApply(observed client.Object, desired client.Object) (*core.ActionCommand, core.CommandResult) {
	act, _ := component.CategoryComponentHandler.PreApply(observed, desired)
	if act.Action == v1.Update {
		// ensure the pod add the cluster ok.
		cmd := component.OrderedReadyPodManagement(observed, desired)
		if cmd != nil {
			act = cmd
		}

		// if the StatefulSet need recreate, ignore the update action.
		recreate := component.RecreateCheck(observed, desired)
		if recreate != nil {
			act = recreate
		}

		// spec change need restart.
		restart := component.RestartCheck(observed, desired)
		if restart != nil {
			act.Append(restart)
		}
	}
	return act, core.Result()
}

// OnEvent make and apply will call it.
func (component *StatefulSetClusterComponent) OnEvent(event extend.Event) error {
	component.Logger().Info("component accept handler event", "category", event.Category, "name", event.Name, "action", event.Action, "state", event.State)
	return nil
}
