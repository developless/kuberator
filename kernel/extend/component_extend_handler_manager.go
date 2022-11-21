package extend

import (
	"github.com/kuberator/api/core"
	"github.com/kuberator/api/extends"
	v1 "github.com/kuberator/api/v1beta1"
	"github.com/kuberator/kernel/common"
)

// CategoryComponentManager handler manager.
// +kubebuilder:object:generate=false
type CategoryComponentManager struct {
	cache map[v1.Category]TypedCategoryComponentHandler
}

// CategoryComponentStageLifeCycleManager handler manager.
// +kubebuilder:object:generate=false
type CategoryComponentStageLifeCycleManager struct {
	cache map[v1.Category]extends.TypedComponentExtendStageLifeCycle
}

func (this CategoryComponentManager) Inject(category v1.Category, component TypedCategoryComponentHandler) CategoryComponentManager {
	this.cache[category] = component
	return this
}

func (this CategoryComponentManager) Get(category v1.Category) TypedCategoryComponentHandler {
	return this.cache[category]
}

func (this CategoryComponentStageLifeCycleManager) Inject(component extends.TypedComponentExtendStageLifeCycle) CategoryComponentStageLifeCycleManager {
	if len(component.GetCategory()) == 0 {
		this.cache[v1.Category(component.GetKind())] = component
		return this
	}
	this.cache[component.GetCategory()] = component
	return this
}

func (this CategoryComponentStageLifeCycleManager) Get(category v1.Category) extends.TypedComponentExtendStageLifeCycle {
	return this.cache[category]
}

var (
	handlerManager         CategoryComponentManager
	CategoryStageLifeCycle CategoryComponentStageLifeCycleManager
)

func init() {
	handlerManager = CategoryComponentManager{cache: map[v1.Category]TypedCategoryComponentHandler{}}
	CategoryStageLifeCycle = CategoryComponentStageLifeCycleManager{cache: map[v1.Category]extends.TypedComponentExtendStageLifeCycle{}}
}

func InjectHandlerIfNotExists(component core.TypedCategoryComponent) TypedCategoryComponentHandler {
	if handlerManager.Get(component.GetCategory()) == nil {
		handler := common.NewTypedObject(component.GetKind()).(TypedCategoryComponentHandler)
		handlerManager.Inject(component.GetCategory(), handler)
		return handler
	}
	return handlerManager.Get(component.GetCategory())
}

func GetHandler(category v1.Category) (TypedCategoryComponentHandler, extends.TypedComponentExtendStageLifeCycle) {
	return handlerManager.Get(category), CategoryStageLifeCycle.Get(category)
}
