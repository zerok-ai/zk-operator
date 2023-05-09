package models

type RestartRequest struct {
	Namespace  string `json:"namespace"`            // This field is mandatory.
	Deployment string `json:"deployment,omitempty"` // This field is mandatory when all field is false.
	All        bool   `json:"all"`                  // Will specify of all deployments need to restarted.
}
