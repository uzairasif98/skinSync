package models

// RBAC models
type Role struct {
	ID          uint64       `gorm:"primaryKey;autoIncrement"`
	Name        string       `gorm:"size:100;uniqueIndex"`
	Description *string      `gorm:"size:255"`
	Permissions []Permission `gorm:"many2many:role_permissions;constraint:OnDelete:CASCADE"`
	Users       []User       `gorm:"many2many:user_roles"`
}

type Permission struct {
	ID          uint64  `gorm:"primaryKey;autoIncrement"`
	Name        string  `gorm:"size:150;uniqueIndex"` // e.g. "users.read", "users.update"
	Description *string `gorm:"size:255"`
	Roles       []Role  `gorm:"many2many:role_permissions"`
}

type RolePermission struct {
	RoleID       uint64 `gorm:"index"`
	PermissionID uint64 `gorm:"index"`
}

type UserRole struct {
	UserID uint64 `gorm:"index"`
	RoleID uint64 `gorm:"index"`
}

// AdminPermission allows per-admin permission overrides
type AdminPermission struct {
	AdminID      uint64     `gorm:"primaryKey;index"`
	PermissionID uint64     `gorm:"primaryKey;index"`
	Granted      bool       `gorm:"default:true"` // true = granted, false = explicitly denied
	Admin        AdminUser  `gorm:"foreignKey:AdminID"`
	Permission   Permission `gorm:"foreignKey:PermissionID"`
}
