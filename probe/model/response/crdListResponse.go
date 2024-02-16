package response

type CRD map[string]interface{}

type CRDListResponse struct {
	CRDList []CRD `json:"crdList"`
}
