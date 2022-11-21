package extend

import (
	"github.com/go-logr/logr"
	"github.com/kuberator/api/core"
	v1 "github.com/kuberator/api/v1beta1"
	. "github.com/kuberator/kernel/common"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Event struct {
	v1.CommonCategoryComponent `json:",inline"`
	Action                     v1.Action             `json:"action,omitempty"`
	Component                  *v1.MiddlewareCluster `json:"component,omitempty"`
	State                      v1.State              `json:"state,omitempty"`
	Result                     core.CommandResult    `json:"result"`
}

// OnEvent action callback
// +kubebuilder:object:generate=false
type OnEvent = func(*Event) error

// TypedCategoryComponentHandler auto defined category component.
// basic component for extend or rewrite.
// +kubebuilder:object:generate=false
type TypedCategoryComponentHandler interface {

	// Make make build-in resource from sub crd.
	// args[0]: the crd meta info.
	// return[0]: the build-in resource desired.
	// return[1]: exception
	Make(core.CustomResource) (*core.ResourcesLine, error)

	// StateFinger provide the cared data finger.
	// args[0]: the desired build-in resource.
	// return[0]: the desired state.
	StateFinger(client.Object) *v1.ComponentState

	// Visitation when the state not change, visitation the relationship resource.
	// args[0]: the context args.
	// return[0]: next stage action.
	Visitation(args core.ComponentArgs) *core.ActionCommand

	// PreApply how to action when apply.
	// args[0]: the observed build-in resource.
	// args[1]: the desired build-in resource.
	// return[0]: next stage action.
	// return[1]: result
	PreApply(client.Object, client.Object) (*core.ActionCommand, core.CommandResult)

	// PostApply post fix the resource running state.
	// args[0]: the action.
	// args[1]: the apply result.
	// return[0]: exception or requeue.
	PostApply(core.ActionCommand, core.CommandResult) core.CommandResult

	// OnEvent lifecycle of component reconcile. .e.g. make and apply will call it.
	OnEvent(Event) error

	// Logger get context logger
	Logger() logr.Logger
}

// CategoryComponentHandler category component handler
type CategoryComponentHandler struct {
}

// Make make the build-in k8s resource from current component crd.
func (component *CategoryComponentHandler) Make(core.CustomResource) (*core.ResourcesLine, error) {
	panic("Not Implement! It must be provide how to make the build-in k8s resource from current component crd")
}

// StateFinger is the sub cr state changed and how to action
func (component *CategoryComponentHandler) StateFinger(client.Object) *v1.ComponentState {
	panic("Not Implement! It must be provide how to generate the state")
}

// Visitation when the state not change, visitation the relationship resource.
func (component *CategoryComponentHandler) Visitation(args core.ComponentArgs) *core.ActionCommand {
	return nil
}

// PreApply how to action when apply.
func (component *CategoryComponentHandler) PreApply(observed client.Object, desired client.Object) (*core.ActionCommand, core.CommandResult) {
	if (observed != nil && desired == nil) || (desired != nil && desired.GetLabels()[ControlLabel] == string(v1.Delete)) {
		target := observed
		if target == nil {
			target = desired
		}
		return &core.ActionCommand{Action: v1.Delete, TargetResource: &core.ReferenceObject{
			Target: target,
		}}, core.Result()
	} else if observed == nil && desired != nil && desired.GetLabels()[ControlLabel] != string(v1.Delete) {
		return &core.ActionCommand{Action: v1.Create,
			TargetResource: &core.ReferenceObject{
				Target: desired,
			}}, core.Result()
	} else if observed != nil && desired != nil {
		desired.SetResourceVersion(observed.GetResourceVersion())
		desired.SetUID(observed.GetUID())
		return &core.ActionCommand{Action: v1.Update, TargetResource: &core.ReferenceObject{
			Target: desired,
		}}, core.Result()
	}
	return &core.ActionCommand{Action: v1.Non, TargetResource: &core.ReferenceObject{}}, core.Result()
}

// PostApply post fix the resource running state.
func (component *CategoryComponentHandler) PostApply(cmd core.ActionCommand, result core.CommandResult) core.CommandResult {
	return result
}

// OnEvent make and apply will call it
func (component *CategoryComponentHandler) OnEvent(Event) error {
	return nil
}

func (component *CategoryComponentHandler) Logger() logr.Logger {
	return ctrl.Log.WithName("CategoryComponentHandler")
}
