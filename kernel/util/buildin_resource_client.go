package util

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/kuberator/api/core"
	v1 "github.com/kuberator/api/v1beta1"
	"github.com/kuberator/kernel/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"time"
)

// ReconcileClient component crd client
type ReconcileClient struct {
	Log    logr.Logger   `json:"log,omitempty"`
	Client client.Client `json:"client,omitempty"`
}

func (cli *ReconcileClient) Get(ctx context.Context, obj client.Object) error {
	return cli.Client.Get(ctx, client.ObjectKeyFromObject(obj), obj)
}

func (cli *ReconcileClient) Create(ctx context.Context, desired client.Object) error {
	err := cli.Client.Create(ctx, desired)
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func (cli *ReconcileClient) UpdateStatus(ctx context.Context, desired client.Object) error {
	return cli.Client.Status().Update(ctx, desired)
}

func (cli *ReconcileClient) Update(ctx context.Context, desired client.Object) error {
	return cli.Client.Update(ctx, desired)
}

func (cli *ReconcileClient) List(ctx context.Context, template metav1.ObjectMeta, object client.ObjectList) error {
	namespace := client.InNamespace(template.GetNamespace())
	matchLabels := client.MatchingLabels(template.GetLabels())
	return client.IgnoreNotFound(cli.Client.List(ctx, object, matchLabels, namespace))
}

func (cli *ReconcileClient) ListPods(ctx context.Context, object client.Object) ([]corev1.Pod, error) {
	var pods corev1.PodList
	err := cli.List(ctx, metav1.ObjectMeta{
		Namespace: object.GetNamespace(),
		Labels:    object.GetLabels(),
	}, &pods)
	return pods.Items, err
}

func (cli *ReconcileClient) Delete(ctx context.Context, observed client.Object) error {
	return client.IgnoreNotFound(cli.Client.Delete(ctx, observed))
}

func (cli *ReconcileClient) DeleteAllOf(ctx context.Context, object *core.ReferenceObject) error {
	var labels client.MatchingLabels
	labels = object.Target.GetLabels()
	return client.IgnoreNotFound(cli.Client.DeleteAllOf(ctx, object.Target, client.InNamespace(object.Target.GetNamespace()), labels))
}

func (reconcile *ReconcileClient) GetIfExists(ctx context.Context, namespace string, source core.TypedCategoryComponent) (client.Object, error) {
	observed := common.NewBuildInResource(source.GetKind(), types.NamespacedName{
		Namespace: namespace,
		Name:      string(source.GetName()),
	})
	err := reconcile.Get(ctx, observed)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			observed = nil
			err = nil
		} else {
			reconcile.Log.Error(err, "get observed build-in resource failed!", "category", source.GetCategory(), "name", source.GetName())
		}
	}
	return observed, err
}

func (cli *ReconcileClient) ReCreate(ctx context.Context, observed client.Object) error {
	reconciledMeta, err := meta.Accessor(observed)
	if err != nil {
		return err
	}

	// Using a precondition here to make sure we delete the version of the resource we intend to delete and
	// to avoid accidentally deleting a resource already recreated for example
	uidToDelete := reconciledMeta.GetUID()
	resourceVersionToDelete := reconciledMeta.GetResourceVersion()

	if len(uidToDelete) == 0 && len(resourceVersionToDelete) == 0 {
		err = cli.Get(ctx, observed)
		if err != nil {
			return err
		}
	}

	propagationPolicy := metav1.DeletePropagationOrphan
	kind := v1.ComponentKind(observed.GetObjectKind().GroupVersionKind().Kind)
	namespaceName := types.NamespacedName{
		Namespace: observed.GetNamespace(),
		Name:      observed.GetName(),
	}

	opts := client.DeleteOptions{
		Preconditions: &metav1.Preconditions{
			UID:             &uidToDelete,
			ResourceVersion: &resourceVersionToDelete,
		},
		PropagationPolicy: &propagationPolicy,
	}

	err = cli.Client.Delete(ctx, observed, &opts)
	if err != nil && !apierrors.IsNotFound(err) {
		cli.Log.Error(err, "delete resource error", "name", observed.GetName())
		return err
	}

	deadLine := time.Now().Add(GetRestartTimeout())
	for true {
		// resourceVersion should not be set on objects to be created.
		observed.SetUID("")
		observed.SetResourceVersion("")
		observed.SetFinalizers(nil)
		err = cli.Client.Create(ctx, observed)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}

		current := common.NewBuildInResource(kind, namespaceName)
		err = cli.Get(ctx, current)
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		} else if len(current.GetUID()) > 0 && current.GetUID() != uidToDelete {
			return nil
		}
		if time.Now().After(deadLine) {
			return errors.New(fmt.Sprintf("(%s/%s) recreate failed!", reconciledMeta.GetNamespace(), reconciledMeta.GetName()))
		}
		cli.Log.Info("waiting the recreate ready", "name", reconciledMeta.GetName())
		time.Sleep(1 * time.Second)
	}

	return err
}

func (cli *ReconcileClient) FailOver(ctx context.Context, observed client.Object, podTemplates ...corev1.Pod) error {
	if len(podTemplates) == 0 {
		return nil
	}

	var failOvers []corev1.Pod

	exists, _, err := cli.CheckIfExists(ctx, podTemplates...)
	if err != nil {
		return err
	}
	if len(podTemplates) == len(exists) {
		for _, pod := range exists {
			if crash := IsPodCrash(pod); crash {
				failOvers = append(failOvers, pod)
			}
		}
	}

	sts := observed.(*appsv1.StatefulSet)
	envMode := GetEnv(sts.Spec.Template.Spec.Containers, "RECOVERY_MODE")
	recovery := false
	if len(failOvers) > 0 && len(failOvers) < len(podTemplates)/2+1 {
		if len(envMode) == 0 || envMode == "false" {
			recovery = true
			// update env set RECOVERY_MODE=true
			ModifyEnv(sts.Spec.Template.Spec.Containers, []corev1.EnvVar{
				{
					Name:  "RECOVERY_MODE",
					Value: "true",
				},
			}...)
			err = cli.Update(ctx, sts)
			if err != nil {
				return err
			}
			cli.Log.Info("update StatefulSet env and set RECOVERY_MODE=true ok.", "name", sts.GetName())
		}
		// sort by start time desc.
		err = cli.Restart(ctx, true, true, failOvers)
		cli.Log.Info("restart crash pod.", "pods", failOvers)
	}

	// avoid restart.
	if recovery || len(failOvers) >= len(podTemplates)/2+1 {
		if GetEnv(sts.Spec.Template.Spec.Containers, "RECOVERY_MODE") == "true" {
			err = cli.Get(ctx, sts)
			// update env set RECOVERY_MODE=true
			ModifyEnv(sts.Spec.Template.Spec.Containers, []corev1.EnvVar{
				{
					Name: "RECOVERY_MODE",
				},
			}...)
			err = cli.Update(ctx, sts)
			cli.Log.Info("update StatefulSet env and reset RECOVERY_MODE ok.", "name", sts.GetName())
		}
	}

	return err
}

func (cli *ReconcileClient) Restart(ctx context.Context, failOver, waitReady bool, pods []corev1.Pod) error {
	timeout := GetRestartTimeout()

	if len(pods) == 0 {
		return errors.New("pod template is empty")
	}

	// validate pods state
	exists, notfound, er := cli.CheckIfExists(ctx, pods...)
	if er != nil {
		return er
	}

	if len(exists) == 0 {
		cli.Log.Info("may all the pods is deleted in other reconcile")
		return nil
	}

	if len(notfound) > 0 {
		time.Sleep(5 * time.Second)
		return errors.New(fmt.Sprintf("exists not found pod"))
	}

	// delete the crash or not ready pod first.
	exists = Ordered(exists...)
	for _, pod := range exists {
		// in order to let the pod move to other health node, fail over need delete the bad pod with it's pvc.
		if failOver {
			// 1. delete pod
			cli.Delete(ctx, &pod)
			deadLine := time.Now().Add(timeout)
			for e := cli.Get(ctx, &pod); e != nil && apierrors.IsNotFound(e); {
				if time.Now().After(deadLine) {
					return errors.New("pod delete failed " + pod.GetName())
				}
				cli.Log.Info("waiting the pod deleted", "name", pod.GetName())
				time.Sleep(3 * time.Second)
			}
			cli.Log.Info("delete pod ok", "name", pod.GetName())

			// 2. delete pvc
			for _, vol := range pod.Spec.Volumes {
				if vol.PersistentVolumeClaim == nil {
					continue
				}
				pvc := &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: pod.Namespace,
						Name:      vol.PersistentVolumeClaim.ClaimName,
					},
				}
				er = cli.Delete(ctx, pvc)
				if er != nil {
					if apierrors.IsNotFound(er) {
						cli.Log.Info("pvc not found, may the pod is deleted in other reconcile", "name", pvc.GetName())
					} else {
						return er
					}
				}

				// 3. wait pvc deleted. avoid pod pending because pvc not exists.
				deadLine = time.Now().Add(timeout)
				for pe := cli.Get(ctx, pvc); pe != nil && apierrors.IsNotFound(er); {
					if time.Now().After(deadLine) {
						return errors.New("pvc delete failed " + vol.PersistentVolumeClaim.ClaimName)
					}
					cli.Log.Info("waiting the pvc deleted", "name", vol.PersistentVolumeClaim.ClaimName)
					time.Sleep(3 * time.Second)
				}

				cli.Log.Info("delete pvc ok", "name", pvc.GetName())
			}
		} else {
			// check is exists pod is not ready, if exists, wait failOver to recovery it and continue.
			if ready := IsPodReady(exists...); !ready {
				time.Sleep(5 * time.Second)
				return errors.New("not all the pods are ready")
			}
		}

		// delete pod
		err := cli.Delete(ctx, &pod)
		if err != nil {
			if apierrors.IsNotFound(err) {
				cli.Log.Info("pod not found, may the pod is deleted in other reconcile", "name", pod.GetName())
			}
			return err
		}

		cli.Log.Info("delete pod ok", "name", pod.GetName())

		deadLine := time.Now().Add(timeout)
		for !cli.TerminalWaitPodReady(ctx, &pod, waitReady, pods...) {
			if time.Now().After(deadLine) {
				return errors.New("pod restart failed " + pod.Name)
			}
			cli.Log.Info("waiting the pod ready", "name", pod.GetName())
			time.Sleep(3 * time.Second)
		}
		cli.Log.Info("pod restart ok", "name", pod.GetName())
	}

	return nil
}

func (cli *ReconcileClient) TerminalWaitPodReady(ctx context.Context, current *corev1.Pod, waitReady bool, templates ...corev1.Pod) bool {
	e, n, er := cli.CheckIfExists(ctx, templates...)
	if er != nil {
		return false
	}
	if len(e) == 0 && len(n) == len(templates) {
		cli.Log.Info("all the pod are deleted", "key", current.GetName())
		return true
	}

	// not exists
	for _, pod := range n {
		if current.Name == pod.Name {
			return false
		}
	}

	// exists
	for _, pod := range e {
		if current.Name == pod.Name {
			if current.UID == pod.UID {
				cli.Log.Info("pod not deleted", "key", pod.GetName())
				return false
			}
			if waitReady {
				return IsPodReady(pod)
			}
			return true
		}
	}

	return true
}

func (cli *ReconcileClient) CheckIfExists(ctx context.Context, templates ...corev1.Pod) ([]corev1.Pod, []corev1.Pod, error) {
	var exists, notfound []corev1.Pod
	for _, template := range templates {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: template.Namespace,
				Name:      template.Name,
			},
		}

		err := cli.Get(ctx, pod)
		if err != nil {
			if apierrors.IsNotFound(err) {
				notfound = append(notfound, template)
				continue
			}
			cli.Log.Error(err, "get pod error", "key", template.Name)
			return exists, notfound, err
		}
		exists = append(exists, *pod)
	}

	return exists, notfound, nil
}

func GetRestartTimeout() time.Duration {
	t := os.Getenv("RESTART_TIMEOUT")
	if len(t) > 0 {
		ot, err := strconv.Atoi(t)
		if err == nil && ot > 0 {
			return time.Duration(ot) * time.Second
		}
	}
	return 3 * time.Minute
}
