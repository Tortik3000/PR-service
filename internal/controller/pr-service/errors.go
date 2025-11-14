package pr_service

import (
	generated "github.com/Tortik3000/PR-service/generated/api/pr-service"
)

func newErrorResponse(code generated.ErrorResponseErrorCode, message string) generated.ErrorResponse {
	return generated.ErrorResponse{
		Error: struct {
			Code    generated.ErrorResponseErrorCode `json:"code"`
			Message string                           `json:"message"`
		}{
			Code:    code,
			Message: message,
		},
	}
}
