package request

// RefreshRequest represents a request to refresh an authentication token.
type RefreshRequest struct {
	RefreshToken string  `json:"refresh_token" validate:"required"`
	DeviceInfo   *string `json:"device_info,omitempty"`
	IPAddress    *string `json:"ip_address,omitempty"`
}
