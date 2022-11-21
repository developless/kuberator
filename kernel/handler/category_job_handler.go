package handler

import (
	"fmt"
	"github.com/kuberator/api/core"
	v1 "github.com/kuberator/api/v1beta1"
	. "github.com/kuberator/kernel/common"
	"github.com/kuberator/kernel/extend"
	. "github.com/kuberator/kernel/util"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Make make the build-in k8s resource from current component crd
func (component *JobHandler) Make(source core.CustomResource) (*core.ResourcesLine, error) {
	// Properties which should be provided from real deployed environment.
	meta := source.ResourceMeta.(*v1.CategoryClusterMixJob)
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: source.Crd.GetNamespace(),
			Name:      string(meta.GetName()),
			OwnerReferences: []metav1.OwnerReference{
				ToOwnerReference(source)},
			Labels:      Merge(source.Crd.GetLabels(), meta.Labels),
			Annotations: Merge(source.Crd.GetLabels(), meta.Labels),
		},
		TypeMeta: metav1.TypeMeta{
			Kind: Job,
		},
		Spec: meta.JobTemplate,
	}

	job.Labels[CategoryLabel] = string(meta.GetCategory())

	return &core.ResourcesLine{
		Desired:      job,
		ResourceMeta: source.ResourceMeta,
	}, nil
}

// StateFinger convert category state to component state
func (component *JobHandler) StateFinger(obj client.Object) *v1.ComponentState {
	if obj == nil {
		return v1.NewComponentState(v1.Deleted, "Deleted", map[string]string{})
	}
	job := obj.(*batchv1.Job)
	data := map[string]string{}
	data["Labels"] = fmt.Sprintf("%v", job.Labels)
	data["Annotation"] = fmt.Sprintf("%v", job.Annotations)
	if job.Spec.Suspend != nil {
		data["template.Suspend"] = fmt.Sprintf("%v", *job.Spec.Suspend)
	}
	if job.Spec.ActiveDeadlineSeconds != nil {
		data["template.ActiveDeadlineSeconds"] = fmt.Sprintf("%v", *job.Spec.ActiveDeadlineSeconds)
	}
	if job.Spec.CompletionMode != nil {
		data["template.CompletionMode"] = fmt.Sprintf("%v", *job.Spec.CompletionMode)
	}
	if job.Spec.ManualSelector != nil {
		data["template.ManualSelector"] = fmt.Sprintf("%v", *job.Spec.ManualSelector)
	}
	if job.Spec.Selector != nil {
		data["template.Selector"] = fmt.Sprintf("%v", *job.Spec.Selector)
	}
	if job.Spec.BackoffLimit != nil {
		data["template.BackoffLimit"] = fmt.Sprintf("%v", *job.Spec.BackoffLimit)
	}
	if job.Spec.ActiveDeadlineSeconds != nil {
		data["template.ActiveDeadlineSeconds"] = fmt.Sprintf("%v", *job.Spec.ActiveDeadlineSeconds)
	}
	if job.Spec.Completions != nil {
		data["template.Completions"] = fmt.Sprintf("%v", *job.Spec.Completions)
	}
	if job.Spec.Parallelism != nil {
		data["template.Parallelism"] = fmt.Sprintf("%v", *job.Spec.Parallelism)
	}
	data["template.template"] = fmt.Sprintf("%v", PodSpecFinger(job.Spec.Template.Spec))
	return v1.NewComponentState(v1.Success, "ok", data)
}

// OnEvent make and apply will call it
func (component *JobHandler) OnEvent(event extend.Event) error {
	component.Logger().Info("configMap accept reconcile event", "event", event)
	return nil
}
