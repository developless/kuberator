package handler

import . "github.com/kuberator/kernel/extend"

type (
	StatefulSetClusterComponent struct {
		CategoryComponentHandler
	}

	ConfigMapComponentHandler struct {
		CategoryComponentHandler
	}

	ServiceComponentHandler struct {
		CategoryComponentHandler
	}

	IngressComponentHandler struct {
		CategoryComponentHandler
	}

	PodDisruptionBudgetHandler struct {
		CategoryComponentHandler
	}

	PersistentVolumeClaimHandler struct {
		CategoryComponentHandler
	}

	SecretHandler struct {
		CategoryComponentHandler
	}

	HorizontalPodAutoscalerHandler struct {
		CategoryComponentHandler
	}

	CronJobHandler struct {
		CategoryComponentHandler
	}

	JobHandler struct {
		CategoryComponentHandler
	}
)
