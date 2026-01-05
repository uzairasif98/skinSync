package response

import "skinSync/models"

// LoginResponse contains tokens and user info
type LoginResponse struct {
	BaseResponse
	Data struct {
		AccessToken      string      `json:"access_token"`
		RefreshToken     string      `json:"refresh_token"`
		AccessExpiresAt  int64       `json:"access_expires_at"`
		RefreshExpiresAt int64       `json:"refresh_expires_at"`
		IsFirstLogin     bool        `json:"is_first_login"`
		User             models.User `json:"user"`
	} `json:"data"`
}
