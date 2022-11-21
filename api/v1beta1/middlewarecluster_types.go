/*
Copyright 2022 wangwei.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	"crypto/md5"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/autoscaling/v2beta2"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sort"
	"strings"
	"time"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MiddlewareClusterSpec defines the desired state of MiddlewareCluster
type MiddlewareClusterSpec struct {
	// cluster version
	// +kubebuilder:validation:Required
	Version string `json:"version,omitempty"`
	// cluster component
	// +kubebuilder:validation:MinItems=1
	Components []*CategoryClusterComponent `json:"components,omitempty"`
	// Component basic properties
	// +optional
	Conf []*NamedProperties `json:"conf,omitempty"`
	// +optional
	Service []*CategoryClusterService `json:"service,omitempty"`
	// +optional
	Ingress []*CategoryClusterIngress `json:"ingress,omitempty"`
	// +optional
	MixJob []*CategoryClusterMixJob `json:"mixJob,omitempty"`
}

func (this MiddlewareClusterSpec) GetVersion() string {
	return this.Version
}

func (this MiddlewareClusterSpec) GetComponents() []*CategoryClusterComponent {
	return this.Components
}

func (this MiddlewareClusterSpec) GetConf() []*NamedProperties {
	return this.Conf
}

func (this MiddlewareClusterSpec) GetService() []*CategoryClusterService {
	return this.Service
}

func (this MiddlewareClusterSpec) GetIngress() []*CategoryClusterIngress {
	return this.Ingress
}

func (this MiddlewareClusterSpec) GetCategoryResource(category Category) interface{} {
	if this.Service != nil {
		for _, svc := range this.Service {
			if svc.GetCategory() == category {
				return svc
			}
		}
	}
	if this.Ingress != nil {
		for _, is := range this.Service {
			if is.GetCategory() == category {
				return is
			}
		}
	}
	if this.MixJob != nil {
		for _, job := range this.MixJob {
			if job.GetCategory() == category {
				return job
			}
		}
	}
	if this.Components != nil {
		for _, c := range this.Components {
			if c.GetCategory() == category {
				return c
			}
		}
	}
	return nil
}

// MiddlewareClusterStatus defines the observed state of MiddlewareCluster
type MiddlewareClusterStatus struct {
	// The state of component.
	// +kubebuilder:validation:Required
	ComponentStatus map[ComponentName]*ComponentState `json:"componentStatus,omitempty"`
	// Uid about the condition for a component.
	// For example, md5 value of the ComponentStatus.
	// +kubebuilder:validation:Required
	Guid string `json:"guid" protobuf:"bytes,2,opt,name=uid,casttype=guid"`
	// UpdateTime about the condition for a component.
	// For example, update time about data.
	// +kubebuilder:validation:Required
	UpdateTimestamp *metav1.Time `json:"updateTimestamp,omitempty" protobuf:"bytes,9,opt,name=updateTimestamp"`
}

func (this *MiddlewareClusterStatus) Init() *MiddlewareClusterStatus {
	this.ComponentStatus = map[ComponentName]*ComponentState{}
	this.Gen()
	return this
}

func NewClusterComponentStatus() *MiddlewareClusterStatus {
	status := &MiddlewareClusterStatus{
		ComponentStatus: map[ComponentName]*ComponentState{},
	}
	return status.Gen()
}

// Gen generator the unique guid
func (this *MiddlewareClusterStatus) Gen() *MiddlewareClusterStatus {
	uidMap := map[string]string{}
	for k, v := range this.ComponentStatus {
		if v != nil {
			uidMap[string(k)] = v.Uid
		}
	}
	data := ToString(uidMap, "=")
	this.Guid = fmt.Sprintf("%x", md5.Sum([]byte(data)))
	this.UpdateTimestamp = &metav1.Time{Time: time.Now()}
	return this
}

// ActionState defines the observed state of ClusterComponent
type ActionState struct {
	// Message about the condition for a component.
	// For example, information about a health check.
	// +kubebuilder:validation:Required
	State State `json:"status" protobuf:"bytes,2,opt,name=status,casttype=status"`
	// Cause about the condition for a component.
	// For example, information about a health check.
	// +optional
	Cause string `json:"cause,omitempty" protobuf:"bytes,3,opt,name=message"`
	// Message about the condition for a component.
	// For example, information about a health check.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,3,opt,name=message"`
	// UpdateTime about the condition for a component.
	// For example, update time about data.
	// +kubebuilder:validation:Required
	UpdateTimestamp *metav1.Time `json:"updateTimestamp,omitempty" protobuf:"bytes,9,opt,name=updateTimestamp"`
}

// ComponentState defines the observed state of ClusterComponent
type ComponentState struct {
	// Uid about the condition for a component.
	// For example, update version about data.
	Uid     string `json:"uid" protobuf:"bytes,2,opt,name=uid,casttype=uid"`
	NextUid string `json:"-"`
	// Details for state
	// +optional
	Details map[string]string `json:"details,omitempty"`
	// Message about the condition for a component.
	// For example, information about a health check.
	// +kubebuilder:validation:Required
	State State `json:"status" protobuf:"bytes,2,opt,name=status,casttype=status"`
	// the reconcile action
	// +optional
	ActionState map[Action]ActionState `json:"actionState"`
	// Message about the condition for a component.
	// For example, information about a health check.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,3,opt,name=message"`
	Meta    string `json:"-"`
	// UpdateTime about the condition for a component.
	// For example, update time about data.
	// +kubebuilder:validation:Required
	UpdateTimestamp *metav1.Time `json:"updateTimestamp,omitempty" protobuf:"bytes,9,opt,name=updateTimestamp"`
}

// IsInProgressAction is action in progress.
func (this *ComponentState) IsInProgressAction() bool {
	if this.ActionState == nil {
		return false
	}
	for _, v := range this.ActionState {
		if v.State != Success && v.State != Failed {
			return true
		}
	}
	return false
}

func (this *ComponentState) IsActionOk() bool {
	if this.ActionState == nil {
		return true
	}
	for _, v := range this.ActionState {
		if v.State != Success {
			return false
		}
	}
	return true
}

// GetActionState generator the unique uid.
func (this *ComponentState) GetActionState(act Action) ActionState {
	if this.ActionState == nil {
		return ActionState{}
	}
	return this.ActionState[act]
}

// UpdateActionState generator the unique uid.
func (this *ComponentState) UpdateActionState(act Action, state State, message string) {
	if this.ActionState == nil {
		this.ActionState = map[Action]ActionState{}
	}
	this.ActionState[act] = ActionState{
		State:           state,
		Message:         message,
		UpdateTimestamp: &metav1.Time{Time: time.Now()},
	}
}

// RecordActionState generator the unique uid.
func (this *ComponentState) RecordActionState(act Action, state State, cause string) {
	if this.ActionState == nil {
		this.ActionState = map[Action]ActionState{}
	}
	this.ActionState[act] = ActionState{
		State:           state,
		Cause:           cause,
		UpdateTimestamp: &metav1.Time{Time: time.Now()},
	}
}

// Gen generator the unique uid.
func (this *ComponentState) Gen(stateFiled map[string]string) *ComponentState {
	if stateFiled == nil {
		stateFiled = map[string]string{}
	}
	this.Meta = ToString(stateFiled, "=")
	this.Uid = fmt.Sprintf("%x", md5.Sum([]byte(this.Meta)))
	return this
}

// NewComponentState ComponentState instance
func NewComponentState(state State, message string, stateFiled map[string]string) *ComponentState {
	var componentState = ComponentState{
		State:           state,
		ActionState:     map[Action]ActionState{},
		Message:         message,
		Details:         map[string]string{},
		UpdateTimestamp: &metav1.Time{Time: time.Now()},
	}
	return componentState.Gen(stateFiled)
}

func ToString(properties map[string]string, separator string) string {
	keys := make([]string, len(properties))
	i := 0
	for k := range properties {
		keys[i] = k
		i = i + 1
	}
	sort.Strings(keys)
	var builder strings.Builder
	for _, key := range keys {
		builder.WriteString(fmt.Sprintf("%s%s%s\n", key, separator, properties[key]))
	}
	return builder.String()
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MiddlewareCluster is the Schema for the middlewareclusters API
type MiddlewareCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MiddlewareClusterSpec   `json:"spec,omitempty"`
	Status MiddlewareClusterStatus `json:"status,omitempty"`
}

func (this *MiddlewareCluster) GetSpec() MiddlewareClusterSpec {
	return this.Spec
}

func (this *MiddlewareCluster) GetStatus() *MiddlewareClusterStatus {
	return &this.Status
}

func (this *MiddlewareCluster) SetStatus(status MiddlewareClusterStatus) {
	this.Status = status
}

func (this *MiddlewareCluster) GetObjectMeta() metav1.ObjectMeta {
	return this.ObjectMeta
}

func (this *MiddlewareCluster) GetTypeMeta() metav1.TypeMeta {
	return this.TypeMeta
}

//+kubebuilder:object:root=true

// MiddlewareClusterList contains a list of MiddlewareCluster
type MiddlewareClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MiddlewareCluster `json:"items"`
}

// NamedProperties named properties
type NamedProperties struct {
	Path string   `json:"path,omitempty"`
	Name string   `json:"name,omitempty"`
	Data string   `json:"data,omitempty"`
	Type ConfType `json:"type,omitempty"`
}

// PropertiesName get properties name
func (p NamedProperties) PropertiesName() string {
	return p.Path + "/" + p.Name
}

// CommonCategoryComponent category component
type CommonCategoryComponent struct {
	Component `json:",inline"`
	// component name
	// +kubebuilder:validation:Required
	Name ComponentName `json:"name,omitempty"`
	// category component role
	// +kubebuilder:validation:Required
	Category Category `json:"category,omitempty"`
	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May match selectors of replication controllers
	// and services.
	// More info: http://kubernetes.io/docs/user-guide/labels
	// +optional
	Labels map[string]string `json:"labels,omitempty" protobuf:"bytes,11,rep,name=labels"`
	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	// More info: http://kubernetes.io/docs/user-guide/annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,12,rep,name=annotations"`
}

// GetCategory category
func (component *CommonCategoryComponent) GetCategory() Category {
	if len(component.Category) == 0 {
		component.SetCategory(Category(fmt.Sprintf("%s-%s", component.GetKind(), component.GetName())))
	}
	return component.Category
}

// GetKind kind
func (component *CommonCategoryComponent) GetKind() ComponentKind {
	return component.Kind
}

// GetName name
func (component *CommonCategoryComponent) GetName() ComponentName {
	return component.Name
}

// GetLabels get labels
func (component *CommonCategoryComponent) GetLabels() map[string]string {
	return component.Labels
}

// GetAnnotations get annotations
func (component *CommonCategoryComponent) GetAnnotations() map[string]string {
	return component.Annotations
}

// SetCategory category
func (component *CommonCategoryComponent) SetCategory(category Category) {
	component.Category = category
}

// SetKind the target build-in kind
func (component *CommonCategoryComponent) SetKind(kind ComponentKind) {
	component.Kind = kind
}

// SetName category name
func (component *CommonCategoryComponent) SetName(name ComponentName) {
	component.Name = name
}

// SetLabels set labels
func (component *CommonCategoryComponent) SetLabels(label map[string]string) {
	component.Labels = labels.Merge(component.Labels, label)
}

// SetAnnotations set annotations
func (component *CommonCategoryComponent) SetAnnotations(annotation map[string]string) {
	component.Annotations = labels.Merge(component.Annotations, annotation)
}

// CategoryClusterComponent basic category component
type CategoryClusterComponent struct {
	CommonCategoryComponent `json:",inline"`
	// properties inject into the component.
	// +optional
	Properties []*NamedProperties `json:"properties,omitempty"`
	// replicas is the desired number of replicas of the given Template, defaults to 1.
	// +kubebuilder:validation:Minimum=0
	Replicas *int32 `json:"replicas,omitempty"`
	// podManagementPolicy controls how pods are created during initial scale up,
	// when replacing pods on nodes, or when scaling down. The default policy is `OrderedReady`
	// The alternative policy is `Parallel` which will create pods in parallel
	// to match the desired scale without waiting, and on scale down will delete all pods at once.
	// +optional
	PodManagementPolicy appsv1.PodManagementPolicyType `json:"podManagementPolicy,omitempty" protobuf:"bytes,6,opt,name=podManagementPolicy,casttype=PodManagementPolicyType"`
	// updateStrategy indicates the StatefulSetUpdateStrategy that will be
	// employed to update Pods in the StatefulSet when a revision is made to
	// Template.
	UpdateStrategy appsv1.StatefulSetUpdateStrategy `json:"updateStrategy,omitempty" protobuf:"bytes,7,opt,name=updateStrategy"`
	// selector is a label query over pods that should match the replica count.
	// It must match the pod template's labels.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	Selector *metav1.LabelSelector `json:"selector"`
	// ServiceName headless name
	ServiceName string `json:"serviceName"`
	// template is the object that describes the pod that will be created
	Template corev1.PodTemplateSpec `json:"template"`
	// volumeClaimTemplates is a list of claims that pods are allowed to reference.
	// +optional
	PersistentVolumeClaim *corev1.PersistentVolumeClaimSpec `json:"persistentVolumeClaim,omitempty"`
	// MaxUnavailable PDB MaxUnavailable.
	// If not setting the effect value, it will work without PDB.
	// +optional
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty" protobuf:"bytes,1,opt,name=maxUnavailable"`
	// maxReplicas is the upper limit for the number of replicas to which the autoscaler can scale up.
	// It cannot be less that minReplicas.
	// +optional
	MaxReplicas *int32 `json:"maxReplicas" protobuf:"varint,3,opt,name=maxReplicas"`
	// metrics contains the specifications for which to use to calculate the
	// desired replica count (the maximum replica count across all metrics will
	// be used).  The desired replica count is calculated multiplying the
	// ratio between the target value and the current value by the current
	// number of pods.  Ergo, metrics used must decrease as the pod count is
	// increased, and vice-versa.  See the individual metric source types for
	// more information about how each type of metric must respond.
	// If not set, the default metric will be set to 80% average CPU utilization.
	// +optional
	Metrics []v2beta2.MetricSpec `json:"metrics,omitempty" protobuf:"bytes,4,rep,name=metrics"`
	// behavior configures the scaling behavior of the target
	// in both Up and Down directions (scaleUp and scaleDown fields respectively).
	// If not set, the default HPAScalingRules for scale up and scale down are used.
	// +optional
	Behavior *v2beta2.HorizontalPodAutoscalerBehavior `json:"behavior,omitempty" protobuf:"bytes,5,opt,name=behavior"`
	// cluster basic auth
	// +optional
	Auth *BasicAuth `json:"auth,omitempty"`
}

func (this *CategoryClusterComponent) GetKind() ComponentKind {
	if len(this.Kind) == 0 {
		this.Kind = "StatefulSet"
	}
	return this.Kind
}

// CategoryClusterService basic category component
type CategoryClusterService struct {
	CommonCategoryComponent `json:",inline"`
	// properties inject into the component.
	corev1.ServiceSpec `json:",inline"`
}

func (this *CategoryClusterService) GetKind() ComponentKind {
	this.Kind = "Service"
	return this.Kind
}

// CategoryClusterIngress basic category component
type CategoryClusterIngress struct {
	CommonCategoryComponent `json:",inline"`
	// properties inject into the component.
	v1beta1.IngressSpec `json:",inline"`
}

func (this *CategoryClusterIngress) GetKind() ComponentKind {
	this.Kind = "Ingress"
	return this.Kind
}

// CategoryClusterMixJob basic category component
type CategoryClusterMixJob struct {
	CommonCategoryComponent `json:",inline"`
	// properties inject into the component.
	// The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron.
	// +optional
	Schedule string `json:"schedule" protobuf:"bytes,1,opt,name=schedule"`

	// Optional deadline in seconds for starting the job if it misses scheduled
	// time for any reason.  Missed jobs executions will be counted as failed ones.
	// +optional
	StartingDeadlineSeconds *int64 `json:"startingDeadlineSeconds,omitempty" protobuf:"varint,2,opt,name=startingDeadlineSeconds"`

	// Specifies how to treat concurrent executions of a Job.
	// Valid values are:
	// - "Allow" (default): allows CronJobs to run concurrently;
	// - "Forbid": forbids concurrent runs, skipping next run if previous run hasn't finished yet;
	// - "Replace": cancels currently running job and replaces it with a new one
	// +optional
	ConcurrencyPolicy batchv1beta1.ConcurrencyPolicy `json:"concurrencyPolicy,omitempty" protobuf:"bytes,3,opt,name=concurrencyPolicy,casttype=ConcurrencyPolicy"`

	// This flag tells the controller to suspend subsequent executions, it does
	// not apply to already started executions.  Defaults to false.
	// +optional
	Suspend *bool `json:"suspend,omitempty" protobuf:"varint,4,opt,name=suspend"`

	// Specifies the job that will be created when executing a CronJob.
	JobTemplate batchv1.JobSpec `json:"jobTemplate" protobuf:"bytes,5,opt,name=jobTemplate"`

	// The number of successful finished jobs to retain.
	// This is a pointer to distinguish between explicit zero and not specified.
	// Defaults to 3.
	// +optional
	SuccessfulJobsHistoryLimit *int32 `json:"successfulJobsHistoryLimit,omitempty" protobuf:"varint,6,opt,name=successfulJobsHistoryLimit"`

	// The number of failed finished jobs to retain.
	// This is a pointer to distinguish between explicit zero and not specified.
	// Defaults to 1.
	// +optional
	FailedJobsHistoryLimit *int32 `json:"failedJobsHistoryLimit,omitempty" protobuf:"varint,7,opt,name=failedJobsHistoryLimit"`
}

func (this *CategoryClusterMixJob) GetKind() ComponentKind {
	if len(this.Kind) == 0 {
		this.Kind = "CronJob"
	}
	return this.Kind
}

func init() {
	SchemeBuilder.Register(&MiddlewareCluster{}, &MiddlewareClusterList{})
}
