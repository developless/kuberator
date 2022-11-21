module github.com/kuberator

go 1.16

require (
	github.com/fatih/structs v1.1.0
	github.com/go-logr/logr v0.4.0
	github.com/imdario/mergo v0.3.13
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/sergi/go-diff v1.2.0
	go.uber.org/zap v1.19.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	sigs.k8s.io/controller-runtime v0.10.0
	sigs.k8s.io/yaml v1.2.0
)
