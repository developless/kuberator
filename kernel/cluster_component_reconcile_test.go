package kernel

import (
	"github.com/kuberator/api/core"
	v1 "github.com/kuberator/api/v1beta1"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
	"strconv"
	"testing"
)

func Crd() (v1.MiddlewareCluster, error) {
	var spec = v1.MiddlewareCluster{}
	content, err := ioutil.ReadFile("../zookeeper/zk-cluster.yaml")
	if err != nil {
		return spec, err
	}
	err = yaml.Unmarshal(content, &spec)
	return spec, err
}

func TestPipeline(t *testing.T) {
	//crd, err := Crd()
	//if err != nil {
	//	t.Error(err)
	//}
	//pipeline := Compile(crd)
	//t.Log("pipeline", *pipeline)
}

func TestClient(t *testing.T) {
	obj := &corev1.Service{}
	version := obj.GetResourceVersion()
	if len(version) > 0 {
		cur, err := strconv.ParseInt(version, 10, 64)
		if err != nil {
			t.Error(err)
		}
		version = strconv.FormatInt(cur+1, 10)
	} else {
		version = "1"
	}
	obj.SetResourceVersion(version)
	t.Log("result", obj.GetResourceVersion())
}

func Get(obj client.Object) {
	obj.SetResourceVersion("666")
	p := &obj
	*p = nil
}

func TestAction(t *testing.T) {
	acts := &core.ActionCommand{
		Action: v1.Create,
		Next: &core.ActionCommand{
			Action: v1.Update,
		},
	}
	acts.Append(&core.ActionCommand{
		Action: v1.Restart,
		Next: &core.ActionCommand{
			Action: v1.Delete,
		},
	})

	for ac := *acts; ; {
		t.Log("ac action", ac.Action)
		if ac.Next == nil {
			break
		}
		ac = *ac.Next
	}

	for a := acts; a != nil; a = a.Next {
		t.Log("action", a.Action)
	}

	for a := acts; a != nil; a = a.Next {
		t.Log("action", a.Action)
	}

	t.Log("action", acts)
}
