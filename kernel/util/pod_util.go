package util

import (
	"fmt"
	"github.com/kuberator/api/core"
	v1 "github.com/kuberator/api/v1beta1"
	. "github.com/kuberator/kernel/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"net"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"strings"
	"time"
)

func TcpGather(pod corev1.Pod) (bool, error) {
	ip := pod.Status.PodIP
	for _, c := range pod.Spec.Containers {
		timeout := c.LivenessProbe.TimeoutSeconds
		port := c.LivenessProbe.TCPSocket.Port
		address := net.JoinHostPort(ip, port.String())
		conn, err := net.DialTimeout("tcp", address, time.Duration(timeout)*time.Second)
		if err != nil {
			if errors.IsTimeout(err) {
				return false, nil
			}
			return false, err
		} else {
			if conn != nil {
				_ = conn.Close()
			} else {
				return false, nil
			}
		}
	}

	return true, nil
}

func OrderedPod(pod types.NamespacedName, no int32) []corev1.Pod {
	var pods []corev1.Pod
	for i := int32(0); i < no; i++ {
		pods = append(pods, corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: pod.Namespace,
				Name:      fmt.Sprintf("%s-%d", pod.Name, i),
			},
		})
	}
	return pods
}

func Ordered(podTemplates ...corev1.Pod) []corev1.Pod {
	var pods []corev1.Pod
	var ready []corev1.Pod
	var failOvers []corev1.Pod
	var notReady []corev1.Pod

	if len(podTemplates) == 0 {
		return podTemplates
	}
	for _, pod := range podTemplates {
		if IsPodCrash(pod) {
			failOvers = append(failOvers, pod)
			continue
		}
		if ok := IsPodReady(pod); !ok {
			notReady = append(notReady, pod)
			continue
		}
		ready = append(ready, pod)
	}

	pods = append(pods, Sort(failOvers...)...)
	pods = append(pods, Sort(notReady...)...)
	pods = append(pods, ready...)

	return pods
}

func Sort(pods ...corev1.Pod) []corev1.Pod {
	if len(pods) > 0 {
		sort.Slice(pods, func(i, j int) bool {
			return pods[i].CreationTimestamp.Before(&pods[j].CreationTimestamp)
		})
	}
	return pods
}

func IsPodReady(pods ...corev1.Pod) bool {
	for _, pod := range pods {
		if pod.Status.Phase != corev1.PodRunning {
			return false
		}
		for _, c := range pod.Status.ContainerStatuses {
			if !c.Ready || c.State.Waiting != nil || c.State.Terminated != nil {
				return false
			}
		}
		for _, c := range pod.Status.Conditions {
			if (c.Type == "Ready" && c.Status != "True") ||
				(c.Type == "ContainersReady" && c.Status != "True") ||
				(c.Type == "PodScheduled" && c.Status != "True") {
				return false
			}
		}
	}
	return true
}

func IsPodCrash(pods ...corev1.Pod) bool {
	for _, pod := range pods {
		startTime := pod.Status.StartTime
		if startTime == nil {
			startTime = &pod.CreationTimestamp
		}
		endTime := startTime.Add(GetRestartTimeout())
		if pod.Status.Phase == corev1.PodFailed ||
			pod.Status.Phase == corev1.PodPhase(v1.Completed) ||
			pod.Status.Phase == corev1.PodPhase(v1.Terminated) ||
			pod.Status.Phase == corev1.PodPhase(v1.Error) ||
			pod.Status.Phase == corev1.PodPending ||
			!IsPodReady(pod) {
			return time.Now().After(endTime)
		}
	}
	return false
}

func GetRestartCommand(desired client.Object, category string, replicas int32, message string) *core.ActionCommand {
	if desired == nil {
		return nil
	}
	labels := desired.GetLabels()
	templatePodName := strings.Join([]string{labels[InstanceLabel], category}, "-")
	return &core.ActionCommand{
		Action:  v1.Restart,
		Message: message,
		TargetResource: &core.ReferenceObject{
			Category: v1.Category(category),
			// template for pod list select
			Target: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: desired.GetNamespace(),
					Name:      templatePodName,
					Labels: map[string]string{
						CategoryLabel: category,
						InstanceLabel: labels[InstanceLabel],
					},
				},
			},
			Extends: OrderedPod(types.NamespacedName{Namespace: desired.GetNamespace(), Name: templatePodName}, replicas),
		},
	}
}

func GetFailOverCommand(desired client.Object) *core.ActionCommand {
	if desired == nil {
		return nil
	}
	if desired.GetObjectKind().GroupVersionKind().Kind != StatefulSet {
		return nil
	}
	labels := desired.GetLabels()
	templatePodName := strings.Join([]string{labels[InstanceLabel], labels[CategoryLabel]}, "-")
	return &core.ActionCommand{
		Action:  v1.FailOver,
		Message: fmt.Sprintf("check if the %s pod need fail over", labels[CategoryLabel]),
		TargetResource: &core.ReferenceObject{
			Category: v1.Category(labels[CategoryLabel]),
			// template for pod list select
			Target:  desired,
			Extends: OrderedPod(types.NamespacedName{Namespace: desired.GetNamespace(), Name: templatePodName}, *desired.(*appsv1.StatefulSet).Spec.Replicas),
		},
	}
}

func GetEnv(templates []corev1.Container, key string) string {
	for _, template := range templates {
		for _, env := range template.Env {
			if env.Name == key {
				return env.Value
			}
		}
	}
	return ""
}

func ModifyEnv(templates []corev1.Container, envVars ...corev1.EnvVar) []corev1.Container {
	envMap := map[string]*corev1.EnvVar{}
	var effectiveEnvs []corev1.EnvVar
	for _, e := range envVars {
		if len(e.Value) > 0 || (e.ValueFrom != nil && e.ValueFrom.Size() > 0) {
			effectiveEnvs = append(effectiveEnvs, e)
		}
		envMap[e.Name] = &e
	}
	for i, c := range templates {
		var envs []corev1.EnvVar
		for _, env := range c.Env {
			if envMap[env.Name] == nil {
				envs = append(envs, env)
			}
		}
		templates[i].Env = append(envs, effectiveEnvs...)
	}

	return templates
}
