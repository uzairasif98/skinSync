package services

import (
	"errors"
	"time"

	"skinSync/config"
	reqdto "skinSync/dto/request"
	resdto "skinSync/dto/response"
	"skinSync/models"

	"gorm.io/gorm"
)

// Login handles unified login/register flow for providers: email, phone, google, apple
func Login(r reqdto.LoginRequest) (*resdto.LoginResponse, error) {
	if r.Provider == "" {
		return nil, errors.New("provider required")
	}

	db := config.DB
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	var user models.User
	var provider models.AuthProvider
	var err error

	switch r.Provider {
	case "email":
		// Email login is handled via OTP in controller
		return nil, errors.New("email login should use OTP flow")

	case "phone":
		if r.Phone == nil {
			return nil, errors.New("phone required")
		}
		if err = db.Where("provider = ? AND phone = ?", "phone", *r.Phone).First(&provider).Error; err == nil {
			if err = db.First(&user, provider.UserID).Error; err != nil {
				return nil, err
			}
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			if err = db.Where("primary_phone = ?", *r.Phone).First(&user).Error; err == nil {
				p := models.AuthProvider{UserID: user.ID, Provider: "phone", Phone: r.Phone}
				if err = db.Create(&p).Error; err != nil {
					return nil, err
				}
				provider = p
			} else {
				user = models.User{PrimaryPhone: r.Phone, Status: "active"}
				if err = db.Create(&user).Error; err != nil {
					return nil, err
				}
				p := models.AuthProvider{UserID: user.ID, Provider: "phone", Phone: r.Phone}
				if err = db.Create(&p).Error; err != nil {
					return nil, err
				}
				provider = p
			}
		} else {
			return nil, err
		}

	case "google":
		if r.GoogleUID == nil {
			return nil, errors.New("google_uid required")
		}
		if err = db.Where("provider = ? AND google_uid = ?", "google", *r.GoogleUID).First(&provider).Error; err == nil {
			if err = db.First(&user, provider.UserID).Error; err != nil {
				return nil, err
			}
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			if r.Email != nil {
				if err = db.Where("primary_email = ?", *r.Email).First(&user).Error; err == nil {
					p := models.AuthProvider{UserID: user.ID, Provider: "google", GoogleUID: r.GoogleUID}
					if err = db.Create(&p).Error; err != nil {
						return nil, err
					}
					provider = p
				} else {
					user = models.User{PrimaryEmail: r.Email, Status: "active"}
					if err = db.Create(&user).Error; err != nil {
						return nil, err
					}
					p := models.AuthProvider{UserID: user.ID, Provider: "google", GoogleUID: r.GoogleUID, Email: r.Email}
					if err = db.Create(&p).Error; err != nil {
						return nil, err
					}
					provider = p
				}
			} else {
				user = models.User{Status: "active"}
				if err = db.Create(&user).Error; err != nil {
					return nil, err
				}
				p := models.AuthProvider{UserID: user.ID, Provider: "google", GoogleUID: r.GoogleUID}
				if err = db.Create(&p).Error; err != nil {
					return nil, err
				}
				provider = p
			}
		} else {
			return nil, err
		}

	case "apple":
		if r.AppleUID == nil {
			return nil, errors.New("apple_uid required")
		}
		if err = db.Where("provider = ? AND apple_uid = ?", "apple", *r.AppleUID).First(&provider).Error; err == nil {
			if err = db.First(&user, provider.UserID).Error; err != nil {
				return nil, err
			}
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			if r.Email != nil {
				if err = db.Where("primary_email = ?", *r.Email).First(&user).Error; err == nil {
					p := models.AuthProvider{UserID: user.ID, Provider: "apple", AppleUID: r.AppleUID}
					if err = db.Create(&p).Error; err != nil {
						return nil, err
					}
					provider = p
				} else {
					user = models.User{PrimaryEmail: r.Email, Status: "active"}
					if err = db.Create(&user).Error; err != nil {
						return nil, err
					}
					p := models.AuthProvider{UserID: user.ID, Provider: "apple", AppleUID: r.AppleUID, Email: r.Email}
					if err = db.Create(&p).Error; err != nil {
						return nil, err
					}
					provider = p
				}
			} else {
				user = models.User{Status: "active"}
				if err = db.Create(&user).Error; err != nil {
					return nil, err
				}
				p := models.AuthProvider{UserID: user.ID, Provider: "apple", AppleUID: r.AppleUID}
				if err = db.Create(&p).Error; err != nil {
					return nil, err
				}
				provider = p
			}
		} else {
			return nil, err
		}

	default:
		return nil, errors.New("invalid provider")
	}

	// create tokens and persist AuthToken
	var emailStr string
	if user.PrimaryEmail != nil {
		emailStr = *user.PrimaryEmail
	}
	token, err := GenerateJWT(emailStr, uint(user.ID), "")
	if err != nil {
		return nil, err
	}
	raw, hashed, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	authToken := models.AuthToken{
		UserID:           user.ID,
		AccessToken:      token,
		RefreshTokenHash: hashed,
		AccessExpiresAt:  time.Now().Add(accessTokenDuration),
		RefreshExpiresAt: time.Now().Add(refreshTokenDuration),
		DeviceInfo:       r.DeviceInfo,
		IPAddress:        r.IPAddress,
	}
	if err = db.Create(&authToken).Error; err != nil {
		return nil, err
	}

	resp := &resdto.LoginResponse{}
	resp.IsSuccess = true
	resp.Message = "Logged in"
	resp.Data.AccessToken = token
	resp.Data.RefreshToken = raw
	resp.Data.AccessExpiresAt = authToken.AccessExpiresAt.Unix()
	resp.Data.RefreshExpiresAt = authToken.RefreshExpiresAt.Unix()
	resp.Data.User = user

	return resp, nil
}

// RefreshToken validates a raw refresh token, rotates it and returns new tokens + user
func RefreshToken(r reqdto.RefreshRequest) (*resdto.LoginResponse, error) {
	if r.RefreshToken == "" {
		return nil, errors.New("refresh_token required")
	}

	db := config.DB
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	// find a non-expired auth token whose hashed refresh token matches the provided raw token
	var tokens []models.AuthToken
	if err := db.Where("refresh_expires_at > ?", time.Now()).Find(&tokens).Error; err != nil {
		return nil, err
	}

	var found *models.AuthToken
	for i := range tokens {
		t := tokens[i]
		// CheckPasswordHash is used elsewhere for bcrypt comparisons
		if CheckPasswordHash(r.RefreshToken, t.RefreshTokenHash) {
			found = &t
			break
		}
	}

	if found == nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	// load user
	var user models.User
	if err := db.First(&user, found.UserID).Error; err != nil {
		return nil, err
	}

	// generate new tokens
	var emailStr string
	if user.PrimaryEmail != nil {
		emailStr = *user.PrimaryEmail
	}
	accessToken, err := GenerateJWT(emailStr, uint(user.ID), "")
	if err != nil {
		return nil, err
	}
	raw, hashed, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// update existing auth token (rotate)
	found.AccessToken = accessToken
	found.RefreshTokenHash = hashed
	found.AccessExpiresAt = time.Now().Add(accessTokenDuration)
	found.RefreshExpiresAt = time.Now().Add(refreshTokenDuration)
	// update optional device/ip info if provided
	if r.DeviceInfo != nil {
		found.DeviceInfo = r.DeviceInfo
	}
	if r.IPAddress != nil {
		found.IPAddress = r.IPAddress
	}

	if err := db.Save(found).Error; err != nil {
		return nil, err
	}

	// prepare response (reusing LoginResponse shape)
	resp := &resdto.LoginResponse{}
	resp.IsSuccess = true
	resp.Message = "Token refreshed"
	resp.Data.AccessToken = accessToken
	resp.Data.RefreshToken = raw
	resp.Data.AccessExpiresAt = found.AccessExpiresAt.Unix()
	resp.Data.RefreshExpiresAt = found.RefreshExpiresAt.Unix()
	resp.Data.User = user

	return resp, nil
}
