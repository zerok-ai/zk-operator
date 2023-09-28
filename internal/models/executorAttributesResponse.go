package models

import zkhttp "github.com/zerok-ai/zk-utils-go/http"

type ExecutorAttributesAPIResponse = zkhttp.ZkHttpResponse[ExecutorAttributesResponse]

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
