package request

// LoginRequest is the request payload for login/register
// Provider can be one of: "email", "phone", "google", "apple"
// For email: require Email and Password
// For phone: require Phone
// For google: require GoogleUID and Email (optional)
// For apple: require AppleUID and Email (optional)

type LoginRequest struct {
	Provider   string  `json:"provider"`
	Email      *string `json:"email,omitempty"`
	Phone      *string `json:"phone,omitempty"`
	Password   *string `json:"password,omitempty"`
	GoogleUID  *string `json:"google_uid,omitempty"`
	AppleUID   *string `json:"apple_uid,omitempty"`
	DeviceInfo *string `json:"device_info,omitempty"`
	IPAddress  *string `json:"ip_address,omitempty"`
}
