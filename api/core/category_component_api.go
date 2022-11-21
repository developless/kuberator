package core

import appsv1beta1 "github.com/kuberator/api/v1beta1"

// TypedCategoryComponent auto defined category component.
// +kubebuilder:object:generate=false
type TypedCategoryComponent interface {
	// GetCategory category name
	GetCategory() appsv1beta1.Category
	// GetKind the target build-in kind
	GetKind() appsv1beta1.ComponentKind
	// GetName get category name
	GetName() appsv1beta1.ComponentName
	// GetLabels get labels
	GetLabels() map[string]string
	// GetAnnotations get annotations
	GetAnnotations() map[string]string
	// SetCategory category name
	SetCategory(appsv1beta1.Category)
	// SetKind the target build-in kind
	SetKind(appsv1beta1.ComponentKind)
	// SetName category name
	SetName(appsv1beta1.ComponentName)
	// SetLabels set labels
	SetLabels(map[string]string)
	// SetAnnotations set annotations
	SetAnnotations(map[string]string)
}
