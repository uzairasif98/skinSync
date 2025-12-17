package request

type UserProfileRequest struct {
	Name            string  `json:"name" validate:"required"`
	PhoneNumber     *string `json:"phone_number,omitempty"`
	EmailAddress    *string `json:"email_address,omitempty" validate:"omitempty,email"`
	Location        *string `json:"location,omitempty"`
	Bio             *string `json:"bio,omitempty"`
	ProfileImageURL *string `json:"profile_image_url,omitempty"`
}
