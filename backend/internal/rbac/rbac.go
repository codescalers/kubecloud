package rbac

import "slices"

type RBACStore interface {
	CreateRole(description string) *Role
	CreatePermission(resource string, action string) *Permission
	CreateAuthorizedUser(userID int) *AuthorizedUser
	GrantRoleToUser(userID int, roleID int) *AuthorizedUser
	RevokeRoleFromUser(userID int, roleID int) *AuthorizedUser
	AddPermissionToRole(roleID int, permissionID int) *Role
	GrantUserPermission(userID int, resource string, resourceID string, action string) *AuthorizedUser
	RevokeUserPermission(userID int, resource string, resourceID string, action string) *AuthorizedUser
}

type Role struct {
	permissions []Permission
	description string
}

type Permission struct {
	resource string
	action   string
}

type Grant struct {
	resource   string
	resourceID string
	action     string
}

type AuthorizedUser struct {
	userID int
	role   []Role
	grants []Grant
}

type RBAC struct {
	roles           []Role
	authorizedUsers []AuthorizedUser
	permissions     []Permission
}

func NewRBAC() *RBAC {
	return &RBAC{
		roles:           []Role{},
		authorizedUsers: []AuthorizedUser{},
		permissions:     []Permission{},
	}
}

func (rbac *RBAC) CreateRole(description string) *Role {
	role := Role{
		description: description,
	}
	rbac.roles = append(rbac.roles, role)
	return &rbac.roles[len(rbac.roles)-1]
}

func (rbac *RBAC) CreatePermission(resource string, action string) *Permission {
	permission := Permission{
		resource: resource,
		action:   action,
	}
	rbac.permissions = append(rbac.permissions, permission)
	return &rbac.permissions[len(rbac.permissions)-1]
}

func (rbac *RBAC) CreateAuthorizedUser(userID int) *AuthorizedUser {
	authorizedUser := AuthorizedUser{
		userID: userID,
		role:   []Role{},
		grants: []Grant{},
	}
	rbac.authorizedUsers = append(rbac.authorizedUsers, authorizedUser)
	return &rbac.authorizedUsers[len(rbac.authorizedUsers)-1]
}

func (rbac *RBAC) GrantRoleToUser(userID int, roleID int) *AuthorizedUser {
	if userID < 0 || userID >= len(rbac.authorizedUsers) {
		return nil
	}
	if roleID < 0 || roleID >= len(rbac.roles) {
		return nil
	}
	authorizedUser := rbac.authorizedUsers[userID]
	authorizedUser.role = append(authorizedUser.role, rbac.roles[roleID])
	rbac.authorizedUsers[userID] = authorizedUser
	return &rbac.authorizedUsers[userID]
}

func (rbac *RBAC) RevokeRoleFromUser(userID int, roleID int) *AuthorizedUser {
	if userID < 0 || userID >= len(rbac.authorizedUsers) {
		return nil
	}
	if roleID < 0 || roleID >= len(rbac.roles) {
		return nil
	}
	authorizedUser := rbac.authorizedUsers[userID]
	target := rbac.roles[roleID]
	authorizedUser.role = slices.DeleteFunc(authorizedUser.role, func(r Role) bool {
		return r.description == target.description
	})
	rbac.authorizedUsers[userID] = authorizedUser
	return &rbac.authorizedUsers[userID]
}

func (rbac *RBAC) AddPermissionToRole(roleID int, permissionID int) *Role {
	if roleID < 0 || roleID >= len(rbac.roles) {
		return nil
	}
	if permissionID < 0 || permissionID >= len(rbac.permissions) {
		return nil
	}
	role := rbac.roles[roleID]
	role.permissions = append(role.permissions, rbac.permissions[permissionID])
	rbac.roles[roleID] = role
	return &rbac.roles[roleID]
}

func (rbac *RBAC) GrantUserPermission(userID int, resource string, resourceID string, action string) *AuthorizedUser {
	if userID < 0 || userID >= len(rbac.authorizedUsers) {
		return nil
	}
	authorizedUser := rbac.authorizedUsers[userID]
	authorizedUser.grants = append(authorizedUser.grants, Grant{resource: resource, resourceID: resourceID, action: action})
	rbac.authorizedUsers[userID] = authorizedUser
	return &rbac.authorizedUsers[userID]
}

func (rbac *RBAC) RevokeUserPermission(userID int, resource string, resourceID string, action string) *AuthorizedUser {
	if userID < 0 || userID >= len(rbac.authorizedUsers) {
		return nil
	}
	authorizedUser := rbac.authorizedUsers[userID]
	authorizedUser.grants = slices.DeleteFunc(authorizedUser.grants, func(g Grant) bool {
		return g.resource == resource && g.resourceID == resourceID && g.action == action
	})
	rbac.authorizedUsers[userID] = authorizedUser
	return &rbac.authorizedUsers[userID]
}

func (rbac *RBAC) Can(userID int, resource string, action string, resourceID ...string) bool {
	if userID < 0 || userID >= len(rbac.authorizedUsers) {
		return false
	}

	id := ""
	if len(resourceID) > 0 {
		id = resourceID[0]
	}

	authorizedUser := rbac.authorizedUsers[userID]

	if id != "" {
		for _, g := range authorizedUser.grants {

			if g.resource != resource {
				continue
			}
			if g.action != action {
				continue
			}
			if g.resourceID != id && g.resourceID != "*" {
				continue
			}
			return true
		}
	}

	for _, role := range authorizedUser.role {
		for _, perm := range role.permissions {
			if perm.resource != resource {
				continue
			}
			if perm.action != action {
				continue
			}
			return true
		}
	}

	return false
}
