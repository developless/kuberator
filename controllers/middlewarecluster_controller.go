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

package controllers

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kuberator/kernel"
	"github.com/kuberator/kernel/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1beta1 "github.com/kuberator/api/v1beta1"
)

// MiddlewareClusterReconciler reconciles a MiddlewareCluster object
type MiddlewareClusterReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Log      logr.Logger
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumes,verbs=get;list;watch
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=policy,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=policy,resources=conjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=extensions,resources=ingresses,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=storage.k8s.io,resources=storageclasses,verbs=get;list;watch
//+kubebuilder:rbac:groups=apps.devless.toplogy.com,resources=middlewareclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.devless.toplogy.com,resources=middlewareclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.devless.toplogy.com,resources=middlewareclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the MiddlewareCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (reconciler *MiddlewareClusterReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	var log = reconciler.Log.WithValues("cluster", request.NamespacedName)

	var reconcile = &kernel.ReconcileContext{
		ReconcileClient: util.ReconcileClient{
			Client: reconciler.Client,
			Log:    reconciler.Log,
		},
		Context:  ctx,
		Request:  request,
		Recorder: reconciler.Recorder,
		Crd: &appsv1beta1.MiddlewareCluster{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: request.Namespace,
				Name:      request.Name,
			},
			Status: *appsv1beta1.NewClusterComponentStatus(),
		},
	}

	result, err := kernel.Reconcile(reconcile).Get()
	if err != nil {
		log.Error(err, "Failed to reconcile")
	}
	if result.RequeueAfter > 0 {
		log.Info("Requeue reconcile request", "after", result.RequeueAfter)
	}

	return result, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *MiddlewareClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1beta1.MiddlewareCluster{}).
		Complete(r)
}
