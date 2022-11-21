package kernel

import (
	"github.com/kuberator/api/core"
	"github.com/kuberator/api/extends"
	v1 "github.com/kuberator/api/v1beta1"
	"github.com/kuberator/kernel/extend"
	"github.com/kuberator/kernel/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func StateFingerStage(reconcile *ReconcileContext, source core.TypedCategoryComponent, observed, desired client.Object) (bool, *v1.ComponentState) {
	ch, sh := extend.GetHandler(source.GetCategory())
	var observedState, desiredState *v1.ComponentState
	var isChanged = false

	// use user define resource finger
	desiredState = finger(ch, sh, desired)
	observedState = finger(ch, sh, observed)

	// 1. resource need create or delete
	// 2. resource update
	// 3. reconcile action not success(.e.g. restart)
	isChanged = ((observed == nil || desired == nil) && observed != desired) ||
		observedState.Uid != desiredState.Uid

	state := reconcile.Crd.GetStatus().ComponentStatus[source.GetName()]
	if state != nil {
		// overwrite the finger data.
		observedState.Details = util.Merge(state.Details, desiredState.Details)
		observedState.ActionState = state.ActionState
		observedState.State = state.State
	}

	if isChanged {
		reconcile.Log.Info("state finger stage framework finger is changed", "category", source.GetCategory(), "name", source.GetName(), "observed", observedState, "desired", desiredState)
		// print diff log.
		util.PrintFingerDiff(observedState.Meta, desiredState.Meta)
		// record next uid. when apply success, use it as the current uid.
		observedState.NextUid = desiredState.Uid
		reconcile.Crd.GetStatus().ComponentStatus[source.GetName()] = observedState
	}

	return isChanged, observedState
}

func finger(ch extend.TypedCategoryComponentHandler, sh extends.TypedComponentExtendStageLifeCycle, target client.Object) *v1.ComponentState {
	var targetState *v1.ComponentState
	if sh != nil {
		state, data := sh.CaredState(target)
		if state != nil {
			targetState = v1.NewComponentState(v1.Success, "ok", state)
			targetState.Details = data
		}
	}

	// use framework handler to finger.
	if targetState == nil && ch != nil {
		targetState = ch.StateFinger(target)
	}

	// avoid nil state.
	if targetState == nil {
		targetState = v1.NewComponentState(v1.Success, "ok", nil)
	}

	return targetState
}
