package response

import "skinSync/models"

// AdminLoginResponse contains admin login data
type AdminLoginResponse struct {
	BaseResponse
	Data struct {
		AccessToken      string           `json:"access_token"`
		RefreshToken     string           `json:"refresh_token"`
		AccessExpiresAt  int64            `json:"access_expires_at"`
		RefreshExpiresAt int64            `json:"refresh_expires_at"`
		AdminUser        models.AdminUser `json:"admin_user"`
	} `json:"data"`
}

// AdminInfoResponse contains admin user info
type AdminInfoResponse struct {
	BaseResponse
	Data models.AdminUser `json:"data"`
}
