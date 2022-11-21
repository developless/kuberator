package core

import (
	appsv1beta1 "github.com/kuberator/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:skip

type BasicCrd interface {
	client.Object
	GetSpec() appsv1beta1.MiddlewareClusterSpec
	GetStatus() *appsv1beta1.MiddlewareClusterStatus
	SetStatus(status appsv1beta1.MiddlewareClusterStatus)
	GetObjectMeta() metav1.ObjectMeta
	GetTypeMeta() metav1.TypeMeta
}

type BasicSpec interface {
	GetVersion() string
	GetComponents() []*appsv1beta1.CategoryClusterComponent
	GetConf() []*appsv1beta1.NamedProperties
	GetService() []*appsv1beta1.CategoryClusterService
	GetIngress() []*appsv1beta1.CategoryClusterIngress
	GetCategoryResource(category appsv1beta1.Category) interface{}
}
