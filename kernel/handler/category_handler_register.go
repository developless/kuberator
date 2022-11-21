package handler

import (
	. "github.com/kuberator/kernel/common"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/autoscaling/v2beta2"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beat1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/api/policy/v1beta1"
)

func init() {
	Inject(StatefulSet, StatefulSetClusterComponent{}, appsv1.StatefulSet{}, appsv1.StatefulSetList{})
	Inject(Service, ServiceComponentHandler{}, corev1.Service{}, corev1.ServiceList{})
	Inject(ConfigMap, ConfigMapComponentHandler{}, corev1.ConfigMap{}, corev1.ConfigMapList{})
	Inject(Ingress, IngressComponentHandler{}, extensionsv1beat1.Ingress{}, extensionsv1beat1.IngressList{})
	Inject(PodDisruptionBudget, PodDisruptionBudgetHandler{}, v1beta1.PodDisruptionBudget{}, v1beta1.PodDisruptionBudgetList{})
	Inject(PersistentVolumeClaim, PersistentVolumeClaimHandler{}, corev1.PersistentVolumeClaim{}, corev1.PersistentVolumeClaimList{})
	Inject(Secret, SecretHandler{}, corev1.Secret{}, corev1.SecretList{})
	Inject(HorizontalPodAutoscaler, HorizontalPodAutoscalerHandler{}, v2beta2.HorizontalPodAutoscaler{}, v2beta2.HorizontalPodAutoscalerList{})
	Inject(CronJob, CronJobHandler{}, batchv1beta1.CronJob{}, batchv1beta1.CronJobList{})
	Inject(Job, JobHandler{}, batchv1.Job{}, batchv1.JobList{})
}
