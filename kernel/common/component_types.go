package common

const (
	StatefulSet             = "StatefulSet"
	Ingress                 = "Ingress"
	Service                 = "Service"
	ConfigMap               = "ConfigMap"
	PodDisruptionBudget     = "PodDisruptionBudget"
	PersistentVolumeClaim   = "PersistentVolumeClaim"
	Secret                  = "Secret"
	HorizontalPodAutoscaler = "HorizontalPodAutoscaler"
	CronJob                 = "CronJob"
	Job                     = "Job"
)

const (
	Category           = "CATEGORY"
	AppName            = "APP_NAME"
	NodeName           = "NODE_NAME"
	HostIp             = "HOST_IP"
	ClusterDomain      = "CLUSTER_DOMAIN"
	PeerService        = "PEER_SVC"
	InitialReplicas    = "INITIAL_CLUSTER_SIZE"
	Namespace          = "NAMESPACE"
	ConfVersion        = "ConfVersion"
	AppConfigMapVolume = "app-config-volume"
	Normal             = "Normal"
	Warning            = "Warning"
)

const (
	ReferenceLabel        = "app.kubernetes.io/reference"
	CategoryLabel         = "app.kubernetes.io/category"
	InstanceLabel         = "app.kubernetes.io/instance"
	AppLabel              = "app.kubernetes.io/app"
	ComponentLabel        = "app.kubernetes.io/component"
	InstancePauseLabel    = "app.kubernetes.io/jd-instance-pause"
	LastAppliedAnnotation = "kubectl.kubernetes.io/last-applied-configuration"
	ControlLabel          = "app.kubernetes.io/control"
)

const (
	ReconcileContextKey = "ReconcileContext"
)
