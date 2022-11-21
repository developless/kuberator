package handler

import (
	"fmt"
	"github.com/kuberator/api/core"
	v1 "github.com/kuberator/api/v1beta1"
	. "github.com/kuberator/kernel/common"
	"github.com/kuberator/kernel/extend"
	. "github.com/kuberator/kernel/util"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Make make the build-in k8s resource from current component crd
func (component *CronJobHandler) Make(source core.CustomResource) (*core.ResourcesLine, error) {
	// Properties which should be provided from real deployed environment.
	meta := source.ResourceMeta.(*v1.CategoryClusterMixJob)
	job := &batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: source.Crd.GetNamespace(),
			Name:      string(meta.GetName()),
			OwnerReferences: []metav1.OwnerReference{
				ToOwnerReference(source)},
			Labels:      Merge(source.Crd.GetLabels(), meta.Labels),
			Annotations: Merge(source.Crd.GetLabels(), meta.Labels),
		},
		TypeMeta: metav1.TypeMeta{
			Kind: CronJob,
		},
		Spec: batchv1beta1.CronJobSpec{
			Schedule:                meta.Schedule,
			StartingDeadlineSeconds: meta.StartingDeadlineSeconds,
			ConcurrencyPolicy:       meta.ConcurrencyPolicy,
			Suspend:                 meta.Suspend,
			JobTemplate: batchv1beta1.JobTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: source.Crd.GetNamespace(),
					Name:      string(meta.GetName()),
					OwnerReferences: []metav1.OwnerReference{
						ToOwnerReference(source)},
					Labels:      Merge(source.Crd.GetLabels(), meta.Labels),
					Annotations: Merge(source.Crd.GetLabels(), meta.Labels),
				},
				Spec: meta.JobTemplate,
			},
			SuccessfulJobsHistoryLimit: meta.SuccessfulJobsHistoryLimit,
			FailedJobsHistoryLimit:     meta.FailedJobsHistoryLimit,
		},
	}

	job.Labels[CategoryLabel] = string(meta.GetCategory())

	return &core.ResourcesLine{
		Desired:      job,
		ResourceMeta: source.ResourceMeta,
	}, nil
}

// StateFinger convert category state to component state
func (component *CronJobHandler) StateFinger(obj client.Object) *v1.ComponentState {
	if obj == nil {
		return v1.NewComponentState(v1.Deleted, "Deleted", map[string]string{})
	}
	job := obj.(*batchv1beta1.CronJob)
	data := map[string]string{}
	data["Labels"] = fmt.Sprintf("%v", job.Labels)
	data["Annotation"] = fmt.Sprintf("%v", job.Annotations)
	data["Schedule"] = fmt.Sprintf("%v", job.Spec.Schedule)
	data["ConcurrencyPolicy"] = fmt.Sprintf("%v", job.Spec.ConcurrencyPolicy)
	if job.Spec.StartingDeadlineSeconds != nil {
		data["StartingDeadlineSeconds"] = fmt.Sprintf("%v", *job.Spec.StartingDeadlineSeconds)
	}
	if job.Spec.Suspend != nil {
		data["Suspend"] = fmt.Sprintf("%v", *job.Spec.Suspend)
	}
	if job.Spec.SuccessfulJobsHistoryLimit != nil {
		data["SuccessfulJobsHistoryLimit"] = fmt.Sprintf("%v", *job.Spec.SuccessfulJobsHistoryLimit)
	}
	if job.Spec.FailedJobsHistoryLimit != nil {
		data["FailedJobsHistoryLimit"] = fmt.Sprintf("%v", *job.Spec.FailedJobsHistoryLimit)
	}
	if job.Spec.JobTemplate.Spec.Suspend != nil {
		data["template.Suspend"] = fmt.Sprintf("%v", *job.Spec.JobTemplate.Spec.Suspend)
	}
	if job.Spec.JobTemplate.Spec.ActiveDeadlineSeconds != nil {
		data["template.ActiveDeadlineSeconds"] = fmt.Sprintf("%v", *job.Spec.JobTemplate.Spec.ActiveDeadlineSeconds)
	}
	if job.Spec.JobTemplate.Spec.CompletionMode != nil {
		data["template.CompletionMode"] = fmt.Sprintf("%v", *job.Spec.JobTemplate.Spec.CompletionMode)
	}
	if job.Spec.JobTemplate.Spec.ManualSelector != nil {
		data["template.ManualSelector"] = fmt.Sprintf("%v", *job.Spec.JobTemplate.Spec.ManualSelector)
	}
	if job.Spec.JobTemplate.Spec.Selector != nil {
		data["template.Selector"] = fmt.Sprintf("%v", *job.Spec.JobTemplate.Spec.Selector)
	}
	if job.Spec.JobTemplate.Spec.BackoffLimit != nil {
		data["template.BackoffLimit"] = fmt.Sprintf("%v", *job.Spec.JobTemplate.Spec.BackoffLimit)
	}
	if job.Spec.JobTemplate.Spec.ActiveDeadlineSeconds != nil {
		data["template.ActiveDeadlineSeconds"] = fmt.Sprintf("%v", *job.Spec.JobTemplate.Spec.ActiveDeadlineSeconds)
	}
	if job.Spec.JobTemplate.Spec.Completions != nil {
		data["template.Completions"] = fmt.Sprintf("%v", *job.Spec.JobTemplate.Spec.Completions)
	}
	if job.Spec.JobTemplate.Spec.Parallelism != nil {
		data["template.Parallelism"] = fmt.Sprintf("%v", *job.Spec.JobTemplate.Spec.Parallelism)
	}
	data["template.template"] = fmt.Sprintf("%v", PodSpecFinger(job.Spec.JobTemplate.Spec.Template.Spec))

	return v1.NewComponentState(v1.Success, "ok", data)
}

// OnEvent make and apply will call it
func (component *CronJobHandler) OnEvent(event extend.Event) error {
	component.Logger().Info("configMap accept reconcile event", "event", event)
	return nil
}
