package resources

type UpdateEnvoyConfRequest struct {
	RateLimit string `json:"rate_limit"`
	Weight    string `json:"weight"`
}

type UpdateEnvoyConfResponse struct {
	IsError      bool            `json:"is_error"`
	ErrorType    string          `json:"error_type"`
	ErrorMessage string          `json:"error_message"`
	ApiResponse  EmptyApiResonse `json:"api_response"`
}

type EmptyApiResonse struct {
}
