package common

type ProgrammingLanguage string

const (
	JavaProgrammingLanguage       ProgrammingLanguage = "java"
	PythonProgrammingLanguage     ProgrammingLanguage = "python"
	GoProgrammingLanguage         ProgrammingLanguage = "go"
	DotNetProgrammingLanguage     ProgrammingLanguage = "dotnet"
	JavascriptProgrammingLanguage ProgrammingLanguage = "javascript"
	UnknownLanguage               ProgrammingLanguage = "unknown"
	NotYetProcessed               ProgrammingLanguage = "notprocessed"
)

type ContainerRuntime struct {
	Image     string            `json:"image"`
	ImageID   string            `json:"imageId"`
	Languages []string          `json:"language"`
	Process   string            `json:"process,omitempty"`
	Cmd       []string          `json:"cmd,omitempty"`
	EnvMap    map[string]string `json:"env"`
}

type RestartRequest struct {
	Namespace  string `json:"namespace"`
	Deployment string `json:"deployment,omitempty"`
	All        bool   `json:"all"`
}
