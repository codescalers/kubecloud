package rbac

import (
	"kubecloud/models"
)

type RBAC struct {
	db models.DB
}

func NewRBAC(db models.DB) *RBAC {
	return &RBAC{db: db}
}

// CanAccess checks if the current user has permission to perform an action on a resource
func (r *RBAC) CanAccess(userID int, resource, action string, resourceID ...string) bool {

	hasRolePermission, err := r.db.CheckUserRolePermission(userID, resource, action)
	if err == nil && hasRolePermission {
		return true
	}

	if len(resourceID) == 0 {
		return false
	}

	hasDirectPermission, err := r.db.CheckUserPermission(userID, resource, action, resourceID[0])
	if err == nil && hasDirectPermission {
		return true
	}

	return false
}
