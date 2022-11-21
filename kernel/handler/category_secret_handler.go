package handler

import (
	"crypto/md5"
	"fmt"
	"github.com/kuberator/api/core"
	v1 "github.com/kuberator/api/v1beta1"
	. "github.com/kuberator/kernel/common"
	. "github.com/kuberator/kernel/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Make make the build-in k8s resource from current component crd
func (component *SecretHandler) Make(source core.CustomResource) (*core.ResourcesLine, error) {
	// Properties which should be provided from real deployed environment.
	meta := source.ResourceMeta.(*core.CategoryComponentObject)
	ref := meta.Reference.(*v1.CategoryClusterComponent)

	if ref == nil || ref.Auth == nil {
		return &core.ResourcesLine{
			ResourceMeta: source.ResourceMeta,
		}, nil
	}

	auth := &v1.BasicAuth{
		Role:     ref.Auth.Role,
		Username: ref.Auth.Username,
		Password: ref.Auth.Password,
		Salt:     ref.Auth.Salt,
		Auth:     ref.Auth.Auth,
	}

	if len(auth.Role) == 0 {
		auth.Role = "root"
	}
	if len(auth.Username) == 0 {
		auth.Username = "root"
	}
	if len(ref.Auth.Password) == 0 {
		auth.Password = fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s_%s_%s_%s", source.Crd.GetNamespace(), source.Crd.GetName(), string(ref.GetCategory()), auth.Salt))))
	}

	template := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: source.Crd.GetNamespace(),
			Name:      string(meta.GetName()),
			OwnerReferences: []metav1.OwnerReference{
				ToOwnerReference(source)},
			Labels:      Merge(nil, source.Crd.GetLabels()),
			Annotations: Merge(nil, source.Crd.GetAnnotations()),
		},
		TypeMeta:   metav1.TypeMeta{Kind: Secret},
		StringData: auth.ToMap(),
		Type:       corev1.SecretTypeBasicAuth,
	}

	template.Labels = Merge(template.Labels, GetReferenceLabels(ref, Secret))

	return &core.ResourcesLine{
		Desired:      template,
		ResourceMeta: source.ResourceMeta,
	}, nil
}

// StateFinger convert category state to component state
func (component *SecretHandler) StateFinger(obj client.Object) *v1.ComponentState {
	if obj == nil {
		return v1.NewComponentState(v1.Deleted, "Deleted", map[string]string{})
	}
	secret := obj.(*corev1.Secret)
	return v1.NewComponentState(v1.Success, "ok", secret.StringData)
}
