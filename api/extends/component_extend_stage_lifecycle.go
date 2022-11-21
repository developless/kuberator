package extends

import (
	"github.com/kuberator/api/core"
	appsv1beta1 "github.com/kuberator/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:skip

type (
	// Callback action callback
	// +kubebuilder:object:generate=false
	Callback = func(interface{}) error

	// TypedComponentExtendStageLifeCycle extend category handler. It will be inject to the base handler.
	// provider auto define extends for developer.
	// +kubebuilder:object:generate=false
	TypedComponentExtendStageLifeCycle interface {
		core.TypedCategoryComponent

		// PostMake post make build-in resource from sub crd
		// args: context arg
		// resourceLine: the desired k8s build-in resource
		// return[0]: the fixed desired k8s build-in resource
		// return[1]: exception
		PostMake(args core.ComponentArgs, resourceLine *core.ResourcesLine) (*core.ResourcesLine, error)

		// CaredState how to construct the state finger.
		// resource: the desired(when create) or the observed k8s(when delete) build-in resource.
		// return[0]: the finger meta data from the resource.
		// return[1]: the data want to save into the state.
		CaredState(resource client.Object) (map[string]string, map[string]string)

		// Visitation when the state not change, visitation the relationship resource.
		// args[0]: the context args.
		// return[0]: next stage action.
		Visitation(args core.ComponentArgs) *core.ActionCommand

		// PreApply pre apply.
		// when the state isn't change, it will be skip.
		// args: context arg
		// observed: the observed(current state) resource.
		// desired: the desired(expect state) resource.
		// return[0]: next action. if Non then, the next stage Apply() not be execute. When action is Restart, it will be tell how to exactly hand it.
		// return[1]: if need requeue and error.
		PreApply(args core.ComponentArgs) (*core.ActionCommand, core.CommandResult)

		// PostApply post apply
		// args: context arg
		// cmd: apply cmd.
		// result: apply result.
		// return: result
		PostApply(args core.ComponentArgs, cmd core.ActionCommand, result core.CommandResult) core.CommandResult
	}
)

type ComponentExtendStageLifeCycle struct {
	appsv1beta1.CommonCategoryComponent
}

func (this ComponentExtendStageLifeCycle) PostMake(args core.ComponentArgs, resourceLine *core.ResourcesLine) (*core.ResourcesLine, error) {
	return nil, nil
}

func (this ComponentExtendStageLifeCycle) CaredState(args client.Object) (map[string]string, map[string]string) {
	return nil, nil
}

func (this ComponentExtendStageLifeCycle) Visitation(args core.ComponentArgs) *core.ActionCommand {
	return nil
}

func (this ComponentExtendStageLifeCycle) PreApply(args core.ComponentArgs) (*core.ActionCommand, core.CommandResult) {
	return nil, core.Result()
}

func (this ComponentExtendStageLifeCycle) PostApply(args core.ComponentArgs, cmd core.ActionCommand, result core.CommandResult) core.CommandResult {
	return result

}
