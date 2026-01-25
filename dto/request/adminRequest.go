package request

// AdminLoginRequest represents admin/clinic login request
type AdminLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// AdminRegisterRequest represents admin/clinic registration request
type AdminRegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required,min=2"`
	RoleName string `json:"role_name" validate:"required"` // super_admin, admin, clinic_manager, clinic_staff
}

// VerifyPasswordRequest represents password verification utility request
type VerifyPasswordRequest struct {
	Password string `json:"password" validate:"required"`
	Hash     string `json:"hash" validate:"required"`
}
