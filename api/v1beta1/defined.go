package v1beta1

const (
	Creating         State = "Creating"
	Reconciling      State = "Reconciling"
	WaitRestart      State = "WaitRestart"
	Restarting       State = "Restarting"
	Updating         State = "Updating"
	Stopping         State = "Stopping"
	PartiallyStopped State = "PartiallyStopped"
	Stopped          State = "Stopped"
	Running          State = "Running"
	Pending          State = "Pending"
	Complicate       State = "Complicate"
	Success          State = "Success"
	Failed           State = "Failed"
	Unknown          State = "Unknown"
	Init             State = "Init"
	Waiting          State = "Waiting"
	Preparing        State = "Preparing"
	InProgress       State = "InProgress"
	NotReady         State = "NotReady"
	Ready            State = "Ready"
	Deleted          State = "Deleted"
	Cancelled        State = "Cancelled"
	Suspended        State = "Suspended"
	Completed        State = "Completed"
	Terminated       State = "Terminated"
	Error            State = "Error"
)

const (
	Create        Action = "Create"
	Update        Action = "Update"
	RollingUpdate Action = "RollingUpdate"
	Scale         Action = "Scale"
	Recycle       Action = "Recycle"
	Delete        Action = "Delete"
	SoftDelete    Action = "SoftDelete"
	Stop          Action = "Stop"
	Start         Action = "Start"
	Restart       Action = "Restart"
	ReCreate      Action = "ReCreate"
	Non           Action = "Non"
	FailOver      Action = "FailOver"
)

const (
	Yaml ConfType = "yaml"
	Json ConfType = "json"
	Ini  ConfType = "ini"
	Text ConfType = "text"
)

type (
	// Category component app role
	Category string

	// ComponentName component name
	ComponentName string

	// ComponentKind build-in resource kind name
	ComponentKind string

	// State of component
	State string

	// Action operator action
	Action string

	// Component the app component
	Component struct {
		// build in resource kind
		// +kubebuilder:validation:Required
		Kind ComponentKind `json:"kind,omitempty"`
	}

	// ConfType conf file type (support: yaml,json,ini,text)
	ConfType string

	BasicAuth struct {
		Role     string `json:"role"`
		Username string `json:"username"`
		Password string `json:"password"`
		Salt     string `json:"salt"`
		Auth     string `json:"auth"`
	}
)

func (this *BasicAuth) ToMap() map[string]string {
	return map[string]string{
		"role":     this.Role,
		"username": this.Username,
		"password": this.Password,
		"auth":     this.Auth,
	}
}
