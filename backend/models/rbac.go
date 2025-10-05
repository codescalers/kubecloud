package models

import "time"

// Role represents a role in the system
type Role struct {
	ID          int          `gorm:"primaryKey;autoIncrement;column:id"`
	Name        string       `gorm:"unique;not null" json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
	Users       []UserRole   `gorm:"foreignKey:RoleID" json:"-"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// Permission represents a permission in the system
type Permission struct {
	ID          int       `gorm:"primaryKey;autoIncrement;column:id"`
	Resource    string    `gorm:"not null" json:"resource"`
	Action      string    `gorm:"not null" json:"action"`
	Description string    `json:"description"`
	Roles       []Role    `gorm:"many2many:role_permissions;" json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserPermission represents direct user permissions (grants)
type UserPermission struct {
	ID         int       `gorm:"primaryKey;autoIncrement;column:id"`
	UserID     int       `gorm:"not null;index" json:"user_id"`
	Resource   string    `gorm:"not null" json:"resource"`
	ResourceID string    `json:"resource_id"`
	Action     string    `gorm:"not null" json:"action"`
	User       User      `gorm:"foreignKey:UserID"`
	CreatedAt  time.Time `json:"created_at"`
}

type UserRole struct {
	UserID    int       `gorm:"primaryKey;column:user_id"`
	RoleID    int       `gorm:"primaryKey;column:role_id"`
	User      User      `gorm:"foreignKey:UserID"`
	Role      Role      `gorm:"foreignKey:RoleID"`
	CreatedAt time.Time `json:"created_at"`
}

// RBAC Methods

// CreateRole creates a new role
func (s *GormDB) CreateRole(role *Role) error {
	return s.db.Create(role).Error
}

// GetRoleByName retrieves a role by name
func (s *GormDB) GetRoleByName(name string) (Role, error) {
	var role Role
	err := s.db.Preload("Permissions").Where("name = ?", name).First(&role).Error
	return role, err
}

// ListRoles lists all roles
func (s *GormDB) ListRoles() ([]Role, error) {
	var roles []Role
	err := s.db.Preload("Permissions").Find(&roles).Error
	return roles, err
}

// UpdateRole updates an existing role
func (s *GormDB) UpdateRole(role *Role) error {
	return s.db.Save(role).Error
}

// DeleteRole deletes a role by ID
func (s *GormDB) DeleteRole(id int) error {
	return s.db.Delete(&Role{}, id).Error
}

// CreatePermission creates a new permission
func (s *GormDB) CreatePermission(permission *Permission) error {
	return s.db.Create(permission).Error
}

// AssignPermissionToRole assigns a permission to a role
func (s *GormDB) AssignPermissionToRole(roleID int, permissionID int) error {
	return s.db.Table("role_permissions").Create(map[string]interface{}{
		"role_id":       roleID,
		"permission_id": permissionID,
	}).Error
}

// RemovePermissionFromRole removes a permission from a role
func (s *GormDB) RemovePermissionFromRole(roleID int, permissionID int) error {
	return s.db.Table("role_permissions").
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(nil).Error
}

// AssignRoleToUser assigns a role to a user
func (s *GormDB) AssignRoleToUser(userID, roleID int) error {
	userRole := &UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	return s.db.Create(userRole).Error
}

// RemoveRoleFromUser removes a role from a user
func (s *GormDB) RemoveRoleFromUser(userID, roleID int) error {
	return s.db.Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&UserRole{}).Error
}

// GetUserRoles gets all roles for a user
func (s *GormDB) GetUserRoles(userID int) ([]Role, error) {
	var roles []Role
	err := s.db.Table("roles").
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Preload("Permissions").
		Find(&roles).Error
	return roles, err
}

// GrantPermissionToUser grants a direct permission to a user
func (s *GormDB) GrantPermissionToUser(userID int, resource, action, resourceID string) error {
	userPermission := &UserPermission{
		UserID:     userID,
		Resource:   resource,
		Action:     action,
		ResourceID: resourceID,
	}
	return s.db.Create(userPermission).Error
}

// RevokePermissionFromUser revokes a direct permission from a user
func (s *GormDB) RevokePermissionFromUser(userID int, resource, action, resourceID string) error {
	return s.db.Where("user_id = ? AND resource = ? AND action = ? AND resource_id = ?",
		userID, resource, action, resourceID).Delete(&UserPermission{}).Error
}

// CheckUserPermission checks if a user has a specific direct permission
func (s *GormDB) CheckUserPermission(userID int, resource, action, resourceID string) (bool, error) {
	var count int64
	err := s.db.Model(&UserPermission{}).
		Where("user_id = ? AND resource = ? AND action = ? AND resource_id = ?",
			userID, resource, action, resourceID).
		Count(&count).Error
	return count > 0, err
}

// CheckUserRolePermission checks if a user has permission through their roles
func (s *GormDB) CheckUserRolePermission(userID int, resource, action string) (bool, error) {
	var count int64
	err := s.db.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
		Where("user_roles.user_id = ? AND permissions.resource = ? AND permissions.action = ?",
			userID, resource, action).
		Count(&count).Error
	return count > 0, err
}
