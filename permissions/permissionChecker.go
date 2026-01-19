package permissions

import (
	"skinSync/config"
	"skinSync/models"
	"sync"
	"time"
)

// CachedPermissions stores permissions with expiry
type CachedPermissions struct {
	Permissions []string
	ExpiresAt   time.Time
}

// In-memory permission cache
var (
	permissionCache = sync.Map{}
	CacheExpiry     = 12 * time.Hour
)

// HasPermission checks if admin has a specific permission (with caching)
func HasPermission(adminID uint64, permissionName string) (bool, error) {
	permissions, err := GetAdminPermissionNames(adminID)
	if err != nil {
		return false, err
	}

	for _, perm := range permissions {
		if perm == permissionName {
			return true, nil
		}
	}
	return false, nil
}

// GetAdminPermissionNames returns all permission names for an admin (with caching)
func GetAdminPermissionNames(adminID uint64) ([]string, error) {
	// Check cache first
	if cached, ok := permissionCache.Load(adminID); ok {
		cachedPerms := cached.(CachedPermissions)
		if time.Now().Before(cachedPerms.ExpiresAt) {
			return cachedPerms.Permissions, nil
		}
		// Cache expired, remove it
		permissionCache.Delete(adminID)
	}

	// Fetch from database
	permissions, err := fetchAdminPermissionsFromDB(adminID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	permissionCache.Store(adminID, CachedPermissions{
		Permissions: permissions,
		ExpiresAt:   time.Now().Add(CacheExpiry),
	})

	return permissions, nil
}

// fetchAdminPermissionsFromDB fetches permissions from database
func fetchAdminPermissionsFromDB(adminID uint64) ([]string, error) {
	db := config.DB
	if db == nil {
		return nil, nil
	}

	// Get admin with role and permissions
	var admin models.AdminUser
	if err := db.Preload("Role.Permissions").First(&admin, adminID).Error; err != nil {
		return nil, err
	}

	// Start with role permissions
	permMap := make(map[string]bool)
	for _, perm := range admin.Role.Permissions {
		permMap[perm.Name] = true
	}

	// Get admin-specific overrides
	var adminPerms []models.AdminPermission
	if err := db.Preload("Permission").Where("admin_id = ?", adminID).Find(&adminPerms).Error; err == nil {
		for _, ap := range adminPerms {
			if ap.Granted {
				// Grant permission
				permMap[ap.Permission.Name] = true
			} else {
				// Deny permission (remove from map)
				delete(permMap, ap.Permission.Name)
			}
		}
	}

	// Convert map to slice
	permissions := make([]string, 0, len(permMap))
	for perm := range permMap {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// InvalidatePermissionCache removes a specific admin's cache
func InvalidatePermissionCache(adminID uint64) {
	permissionCache.Delete(adminID)
}

// ClearAllPermissionCache clears the entire permission cache
func ClearAllPermissionCache() {
	permissionCache.Range(func(key, value interface{}) bool {
		permissionCache.Delete(key)
		return true
	})
}

// GetAdminPermissionsGrouped returns permissions grouped by category
func GetAdminPermissionsGrouped(adminID uint64) (map[string][]models.Permission, error) {
	db := config.DB
	if db == nil {
		return nil, nil
	}

	// Get admin with role and permissions
	var admin models.AdminUser
	if err := db.Preload("Role.Permissions").First(&admin, adminID).Error; err != nil {
		return nil, err
	}

	// Start with role permissions
	permMap := make(map[string]models.Permission)
	for _, perm := range admin.Role.Permissions {
		permMap[perm.Name] = perm
	}

	// Apply admin-specific overrides
	var adminPerms []models.AdminPermission
	if err := db.Preload("Permission").Where("admin_id = ?", adminID).Find(&adminPerms).Error; err == nil {
		for _, ap := range adminPerms {
			if ap.Granted {
				permMap[ap.Permission.Name] = ap.Permission
			} else {
				delete(permMap, ap.Permission.Name)
			}
		}
	}

	// Group by first part of permission name (e.g., "users.view" -> "users")
	grouped := make(map[string][]models.Permission)
	for _, perm := range permMap {
		// Extract category from permission name (e.g., "users.view" -> "users")
		category := "general"
		if len(perm.Name) > 0 {
			for i, c := range perm.Name {
				if c == '.' || c == '_' {
					category = perm.Name[:i]
					break
				}
			}
		}
		grouped[category] = append(grouped[category], perm)
	}

	return grouped, nil
}
