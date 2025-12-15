package response

type BaseResponse struct {
	IsSuccess bool   `json:"is_success"`
	Message   string `json:"message"`
}
