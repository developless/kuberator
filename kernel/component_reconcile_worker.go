package kernel

import (
	"fmt"
	"github.com/kuberator/api/core"
	v1 "github.com/kuberator/api/v1beta1"
	. "github.com/kuberator/kernel/common"
	"github.com/kuberator/kernel/extend"
	"github.com/kuberator/kernel/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type (
	HandlerFunc     = func(core.TypedCategoryComponent) error
	MakeFunc        = func(reconcile *ReconcileContext, task core.TypedCategoryComponent) (*core.ResourcesLine, error)
	MergeFunc       = func(reconcile *ReconcileContext, resource *core.ResourcesLine) error
	StateFingerFunc = func(reconcile *ReconcileContext, task core.TypedCategoryComponent, observed, desired client.Object) (bool, *v1.ComponentState)
	VisitationFunc  = func(reconcile *ReconcileContext, task core.TypedCategoryComponent, observed, desired client.Object) *core.ActionCommand
	PreApplyFunc    = func(reconcile *ReconcileContext, task core.TypedCategoryComponent, observed, desired client.Object) (*core.ActionCommand, core.CommandResult)
	ApplyFunc       = func(reconcile *ReconcileContext, command *core.ActionCommand) core.CommandResult
	PostApplyFunc   = func(reconcile *ReconcileContext, command core.ActionCommand, result core.CommandResult) core.CommandResult
	ReduceFunc      = func(reconcile *ReconcileContext, result core.CommandResult) core.CommandResult

	Pipeline struct {
		reconcile     *ReconcileContext
		chain         []core.TypedCategoryComponent
		make          MakeFunc
		merge         MergeFunc
		stateFinger   StateFingerFunc
		visitation    VisitationFunc
		preApply      PreApplyFunc
		apply         ApplyFunc
		postApply     PostApplyFunc
		reduce        ReduceFunc
		ResourcesLine *core.ResourcesLine
		ActionCommand *core.ActionCommand
	}
)

func (this *Pipeline) add(task core.TypedCategoryComponent) *Pipeline {
	if task == nil {
		return this
	}
	this.chain = append(this.chain, task)
	extend.InjectHandlerIfNotExists(task)
	return this
}

func (this *Pipeline) WithMakeFunc(fun MakeFunc) *Pipeline {
	this.make = fun
	return this
}

func (this *Pipeline) WithMergeFunc(fun MergeFunc) *Pipeline {
	this.merge = fun
	return this
}

func (this *Pipeline) WithStateFingerFunc(fun StateFingerFunc) *Pipeline {
	this.stateFinger = fun
	return this
}

func (this *Pipeline) WithVisitationFunc(fun VisitationFunc) *Pipeline {
	this.visitation = fun
	return this
}

func (this *Pipeline) WithPreApplyFunc(fun PreApplyFunc) *Pipeline {
	this.preApply = fun
	return this
}

func (this *Pipeline) WithApplyFunc(fun ApplyFunc) *Pipeline {
	this.apply = fun
	return this
}

func (this *Pipeline) WithPostApplyFunc(fun PostApplyFunc) *Pipeline {
	this.postApply = fun
	return this
}

func (this *Pipeline) WithReduceFunc(fun ReduceFunc) *Pipeline {
	this.reduce = fun
	return this
}

func (this *Pipeline) resourcePipeline() error {
	for _, task := range this.chain {
		command, err := this.make(this.reconcile, task)
		if err != nil {
			return err
		}
		if command == nil {
			continue
		}
		if this.ResourcesLine == nil {
			this.ResourcesLine = command
		} else {
			this.ResourcesLine.Append(command)
		}
	}
	return nil
}

func (this *Pipeline) actionPipeline() core.CommandResult {
	restartMap := map[v1.Category]*core.ActionCommand{}
	skipRestartMap := map[v1.Category]bool{}
	for cmd := this.ResourcesLine; cmd != nil; cmd = cmd.Next {
		var action *core.ActionCommand
		result := core.Result()

		this.merge(this.reconcile, cmd)
		isChanged, state := this.stateFinger(this.reconcile, cmd.ResourceMeta, cmd.Observed, cmd.Desired)

		if isChanged {
			action, result = this.preApply(this.reconcile, cmd.ResourceMeta, cmd.Observed, cmd.Desired)
		} else {
			action = this.visitation(this.reconcile, cmd.ResourceMeta, cmd.Observed, cmd.Desired)
			// check unknown state action.
			if !state.IsActionOk() {
				// wait restart
				if cmd.Desired != nil && len(state.ActionState[v1.Restart].State) > 0 && state.ActionState[v1.Restart].State != v1.Success {
					labels := cmd.Desired.GetLabels()
					category := labels[ReferenceLabel]
					if len(category) == 0 {
						category = labels[CategoryLabel]
					}
					act := util.GetRestartCommand(cmd.Desired, category, 0, state.Message)
					if action == nil {
						action = act
					} else {
						action.Append(act)
					}
				}
			}
		}

		if result.NotEmpty() {
			return result
		}

		if action == nil {
			continue
		}

		for a := action; a != nil; a = a.Next {
			if a.ResourceMeta == nil {
				a.ResourceMeta = cmd.ResourceMeta
			}
			if len(a.TargetResource.Category) == 0 {
				a.TargetResource.Category = cmd.ResourceMeta.GetCategory()
			}

			node := *a
			node.Next = nil

			state.RecordActionState(a.Action, v1.Preparing, a.Message)
			// skip restart.
			if a.Action == v1.Create || a.Action == v1.Delete {
				skipRestartMap[a.TargetResource.Category] = true
			}
			if a.Action == v1.Restart {
				state.RecordActionState(a.Action, v1.WaitRestart, a.Message)
				restartMap[node.TargetResource.Category] = &node
				continue
			}

			// add action line.
			if this.ActionCommand == nil {
				this.ActionCommand = &node
			} else {
				this.ActionCommand.MoveIfAbsent(&node)
			}
		}
		this.reconcile.Crd.GetStatus().ComponentStatus[cmd.ResourceMeta.GetName()] = state
	}

	// append restart command.
	for k, v := range restartMap {
		cause := "UNKNOWN"
		if len(v.Message) > 0 {
			cause = v.Message
		}
		if skipRestartMap[k] {
			this.reconcile.Log.Info("restart", "category", v.TargetResource.Category, "name", v.ResourceMeta.GetName(), "cause", cause, "result", "skip")
			continue
		}
		// add action line.
		if this.ActionCommand == nil {
			this.ActionCommand = v
		} else {
			this.ActionCommand.Append(v)
		}
	}

	return core.Result()
}

func (this *Pipeline) exec() core.CommandResult {
	result := core.Result()
	for cmd := this.ActionCommand; !result.NotEmpty() && cmd != nil; cmd = cmd.Next {
		state := this.reconcile.Crd.GetStatus().ComponentStatus[cmd.ResourceMeta.GetName()]
		if state == nil {
			this.reconcile.Log.Info("state not found", "category", cmd.ResourceMeta.GetCategory(), "resource name", cmd.ResourceMeta.GetName())
			state = v1.NewComponentState(v1.Success, "unknown state", nil)
		}

		// not update the status.
		state.UpdateActionState(cmd.Action, v1.InProgress, cmd.Message)
		this.reconcile.Log.Info("apply stage", "action", cmd.Action, "category", cmd.TargetResource.Category, "name", cmd.ResourceMeta.GetName())

		if cmd.Validate != nil {
			result.Error(cmd.Validate(this.reconcile.Client))
		}

		result = this.apply(this.reconcile, cmd)
		if result.IsError() {
			state.UpdateActionState(cmd.Action, v1.Failed, result.LastError().Error())
			this.reconcile.Recorder.Eventf(cmd.TargetResource.Target, Normal, string(cmd.Action), result.LastError().Error())
		} else {
			this.reconcile.Log.Info("apply success", "action", cmd.Action, "category", cmd.TargetResource.Category, "name", cmd.ResourceMeta.GetName())
			// update state
			state.UpdateActionState(cmd.Action, v1.Success, "")
			if len(state.NextUid) > 0 {
				state.Uid = state.NextUid
			}
		}

		this.reconcile.Crd.GetStatus().ComponentStatus[cmd.ResourceMeta.GetName()] = state

		if cmd.Callback != nil {
			result.Error(cmd.Callback(&result, this.reconcile.Client))
		}

		result.Print(this.reconcile.Log)
		// post apply
		this.reconcile.Log.Info("post apply stage", "action", cmd.Action, "category", cmd.TargetResource.Category, "name", cmd.ResourceMeta.GetName())
		result.Merge(PostApplyStage(this.reconcile, *cmd, result))
	}

	return result
}

func (this *Pipeline) Compute() core.CommandResult {
	this.reconcile.Log.Info("begin build in resources make stage")
	if err := this.resourcePipeline(); err != nil {
		this.reconcile.Log.Info("build in resources make stage failed.")
		return core.Result().Error(err)
	}
	this.reconcile.Log.Info("build in resources make ok, begin command construct stage.")
	if r := this.actionPipeline(); r.NotEmpty() {
		this.reconcile.Log.Info("command construct stage error or exit, begin reduce stage.")
		return this.reduce(this.reconcile, r)
	}
	if this.ActionCommand == nil {
		this.reconcile.Log.Info("action stage exit with no command.")
		return core.Result()
	}
	this.reconcile.Log.Info("command construct stage ok, begin exec action stage.")
	result := this.exec()
	result.Print(this.reconcile.Log)
	this.reconcile.Log.Info("begin reduce stage.")
	return this.reduce(this.reconcile, result)
}

func Compile(reconcile *ReconcileContext) *Pipeline {
	pipeline := &Pipeline{reconcile: reconcile}
	crd := reconcile.Crd
	// headless service need add to it.
	for _, task := range crd.GetSpec().Service {
		pipeline.add(Format(task, crd))
	}
	cms := util.BuildConfResource(crd.GetSpec().Conf, crd.GetSpec().Components)
	for _, task := range cms {
		pipeline.add(Format(task, crd))
	}
	for _, task := range crd.GetSpec().Components {
		pipeline.add(Format(task, crd))
		//PVC
		pipeline.add(Format(InferResource(task, PersistentVolumeClaim), crd))
		//HPA
		pipeline.add(Format(InferResource(task, HorizontalPodAutoscaler), crd))
		//PDB
		pipeline.add(Format(InferResource(task, PodDisruptionBudget), crd))
		//Secret
		pipeline.add(Format(InferResource(task, Secret), crd))
	}
	for _, task := range crd.GetSpec().Ingress {
		pipeline.add(Format(task, crd))
	}

	for _, task := range crd.GetSpec().MixJob {
		pipeline.add(Format(task, crd))
	}
	return pipeline
}

func InferResource(component *v1.CategoryClusterComponent, kind v1.ComponentKind) *core.CategoryComponentObject {
	return &core.CategoryComponentObject{
		CommonCategoryComponent: v1.CommonCategoryComponent{
			Category: v1.Category(strings.ToLower(fmt.Sprintf("%s-%s", component.GetCategory(), kind))),
			Component: v1.Component{
				Kind: kind,
			},
		},
		Reference: component,
	}
}

func Format(task core.TypedCategoryComponent, com core.BasicCrd) core.TypedCategoryComponent {
	if task == nil {
		return nil
	}
	if len(task.GetName()) == 0 {
		task.SetName(v1.ComponentName(task.GetCategory()))
	}
	if !strings.HasPrefix(string(task.GetName()), com.GetName()) {
		task.SetName(v1.ComponentName(util.GetComponentShotName(com.GetName(), v1.Category(task.GetName()))))
	}
	extend.InjectHandlerIfNotExists(task)
	return task
}
