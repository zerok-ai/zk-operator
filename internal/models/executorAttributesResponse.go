package models

type ExecutorAttributesResponse struct {
	ExecutorAttributes []ExecutorVersionAttrSet `json:"executor_attributes"`
	Version            int64                    `json:"version"`
	Update             bool                     `json:"update"`
}

type ExecutorVersionAttrSet struct {
	Executor   string                 `json:"executor"`
	Version    string                 `json:"version"`
	Protocol   string                 `json:"protocol"`
	Attributes map[string]interface{} `json:"attributes"`
}
