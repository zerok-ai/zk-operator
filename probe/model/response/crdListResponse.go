package response

import (
	"github.com/zerok-ai/zk-utils-go/scenario/model"
)

type CRDListResponse struct {
	CRDList []model.Scenario `json:"crdList"`
}
