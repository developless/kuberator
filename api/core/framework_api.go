package core

import (
	"context"
	"github.com/go-logr/logr"
	appsv1beta1 "github.com/kuberator/api/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type (
	// Callback action callback
	// +kubebuilder:object:generate=false
	Callback = func(*CommandResult, client.Client, ...interface{}) error

	// Validate action validate
	// +kubebuilder:object:generate=false
	Validate = func(client.Client, ...interface{}) error

	// CustomResource source resource
	CustomResource struct {
		ResourceMeta TypedCategoryComponent `json:"resourceMeta,omitempty"`
		Crd          BasicCrd               `json:"crd,omitempty"`
	}

	// ResourcesLine resource collection
	ResourcesLine struct {
		ResourceMeta TypedCategoryComponent `json:"resourceMeta,omitempty"`
		Observed     client.Object          `json:"observed,omitempty"`
		Desired      client.Object          `json:"desired,omitempty"`
		Next         *ResourcesLine         `json:"next,omitempty"`
	}

	// ActionCommand action command
	ActionCommand struct {
		Action         appsv1beta1.Action     `json:"action,omitempty"`
		Message        string                 `json:"message"`
		TargetResource *ReferenceObject       `json:"targetResource"`
		ResourceMeta   TypedCategoryComponent `json:"resourceMeta,omitempty"`
		Next           *ActionCommand         `json:"next"`
		Callback       Callback
		Validate       Validate
	}

	// ReferenceObject reference object.
	ReferenceObject struct {
		// target resource category
		Category appsv1beta1.Category `json:"category,omitempty"`
		// target build-in object template pointer. if restart/pvc/pdb operator, it is the select template with select labels.
		Target client.Object `json:"target,omitempty"`
		// the extends args
		Extends interface{} `json:"extends,omitempty"`
	}

	// CategoryComponentObject category object.
	CategoryComponentObject struct {
		appsv1beta1.CommonCategoryComponent `json:",inline"`
		Object                              interface{}
		Reference                           TypedCategoryComponent
	}

	// ComponentArgs component context
	ComponentArgs struct {
		Context              context.Context `json:"context,omitempty"`
		CustomResource       `json:",inline"`
		types.NamespacedName `json:",inline"`
		Observed             client.Object `json:"observed,omitempty"`
		Desired              client.Object `json:"desired,omitempty"`
		Logger               logr.Logger
	}

	// CommandResult command result
	CommandResult struct {
		delete bool
		result *ctrl.Result
		errors []error
	}
)

func Result() CommandResult {
	return CommandResult{
		delete: false,
		result: &ctrl.Result{},
		errors: []error{},
	}
}

func (this CommandResult) Print(log logr.Logger) {
	if log == nil {
		log = ctrl.Log.WithName("CommandResult")
	}
	this.Get()
	log.Info("execute result", "requeue", this.result.Requeue, "requeueAfter", this.result.RequeueAfter, "errors", len(this.errors))
	for i, err := range this.errors {
		if err == nil {
			continue
		}
		log.Error(err, "result errors", "index", i)
	}
}

func (this CommandResult) WithDelete() CommandResult {
	this.delete = true
	return this
}

func (this CommandResult) WithRequeue() CommandResult {
	this.result.Requeue = true
	return this
}

func (this CommandResult) WithRequeueAfter(requeueAfter time.Duration) CommandResult {
	this.result.RequeueAfter = requeueAfter
	return this
}

func (this CommandResult) Error(err ...error) CommandResult {
	if err == nil || len(err) == 0 {
		return this
	}
	for _, e := range err {
		if e != nil {
			this.errors = append(this.errors, e)
		}
	}
	return this
}

func (this CommandResult) Merge(result CommandResult) CommandResult {
	this.Error(result.errors...)
	if !this.NotEmpty() || result.NotEmpty() {
		if result.result != nil {
			this.result = result.result
		}
		return this
	}
	return this
}

func (this CommandResult) NotEmpty() bool {
	return this.IsError() || this.result.Requeue || this.result.RequeueAfter > 0
}

func (this CommandResult) IsError() bool {
	return this.errors != nil && len(this.errors) > 0
}

func (this CommandResult) LastError() error {
	if this.IsError() {
		return this.errors[len(this.errors)-1]
	}
	return nil
}

func (this CommandResult) Fire(requeueAfter time.Duration) (ctrl.Result, error) {
	if this.errors != nil && len(this.errors) > 0 {
		for _, err := range this.errors {
			if apierrors.IsAlreadyExists(err) || apierrors.IsConflict(err) || apierrors.IsNotFound(err) {
				this.result.Requeue = true
				return *this.result, nil
			}
		}
		return *this.result, this.errors[0]
	}
	if this.result == nil {
		this.result = &ctrl.Result{}
	}
	if !this.delete && !this.result.Requeue && this.result.RequeueAfter == 0 && requeueAfter > 0 {
		this.result.RequeueAfter = requeueAfter
	}
	return *this.result, nil
}

func (this CommandResult) Get() (ctrl.Result, error) {
	return this.Fire(0)
}

func (this *ResourcesLine) Append(command *ResourcesLine) *ResourcesLine {
	if command == nil {
		return this
	}
	for cmd := this; cmd != nil; cmd = cmd.Next {
		if cmd.Next == nil {
			cmd.Next = command
			return this
		}
	}
	return this
}

func (this *ActionCommand) Append(command *ActionCommand) *ActionCommand {
	if command == nil {
		return this
	}
	for cmd := this; cmd != nil; cmd = cmd.Next {
		if cmd.Next == nil {
			cmd.Next = command
			return this
		}
	}
	return this
}

func (this *ActionCommand) MoveIfAbsent(command *ActionCommand) *ActionCommand {
	if command == nil {
		return this
	}
	head := this
	for cmd := this; cmd != nil; cmd = cmd.Next {
		// remove the same action.
		if cmd.Action == command.Action && cmd.TargetResource.Category == command.TargetResource.Category {
			cmd = cmd.Next
			if cmd == nil {
				head.Next = command
				return this
			}
			head.Next = cmd
		}
		if cmd.Next == nil {
			cmd.Next = command
			return this
		}
		head = cmd
	}
	return this
}
