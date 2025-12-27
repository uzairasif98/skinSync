package request

// SendOTPRequest - request to send OTP to email
type SendOTPRequest struct {
	Email      string  `json:"email" validate:"required,email"`
	DeviceInfo *string `json:"device_info,omitempty"`
	IPAddress  *string `json:"ip_address,omitempty"`
}

// VerifyOTPRequest - request to verify OTP and login
type VerifyOTPRequest struct {
	Email      string  `json:"email" validate:"required,email"`
	OTP        string  `json:"otp" validate:"required,len=6"`
	DeviceInfo *string `json:"device_info,omitempty"`
	IPAddress  *string `json:"ip_address,omitempty"`
}
