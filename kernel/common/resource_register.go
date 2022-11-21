package common

import (
	v1 "github.com/kuberator/api/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	typedMap, buildInMap, buildInList map[v1.ComponentKind]reflect.Type
)

func init() {
	typedMap = map[v1.ComponentKind]reflect.Type{}
	buildInMap = map[v1.ComponentKind]reflect.Type{}
	buildInList = map[v1.ComponentKind]reflect.Type{}
}

func Inject(kind v1.ComponentKind, handler, buildIn, buildIns interface{}) {
	typedMap[kind] = reflect.TypeOf(handler)
	buildInMap[kind] = reflect.TypeOf(buildIn)
	buildInList[kind] = reflect.TypeOf(buildIns)
}

func NewTypedObject(kind v1.ComponentKind) interface{} {
	return reflect.New(typedMap[kind]).Interface()
}

func NewBuildInResource(kind v1.ComponentKind, namespaceName types.NamespacedName) client.Object {
	obj := reflect.New(buildInMap[kind]).Interface().(client.Object)
	if obj != nil {
		gvk := obj.GetObjectKind().GroupVersionKind()
		gvk.Kind = string(kind)
		obj.GetObjectKind().SetGroupVersionKind(gvk)
		obj.SetNamespace(namespaceName.Namespace)
		obj.SetName(namespaceName.Name)
	}
	return obj
}

func NewBuildInListResource(kind v1.ComponentKind) client.ObjectList {
	return reflect.New(buildInList[kind]).Interface().(client.ObjectList)
}
