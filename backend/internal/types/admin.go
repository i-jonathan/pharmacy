package types

import "time"

type AdminModuleInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Path        string `json:"path"`
}

// CreateUserRequest is an admin adding a staff account with its role and starting password.
type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	RoleID   int    `json:"role_id"`
}

type PermissionInfo struct {
	ID       int    `json:"id"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

type RoleWithPermissions struct {
	ID          int              `json:"id"`
	Name        string           `json:"name"`
	Permissions []PermissionInfo `json:"permissions"`
	CreatedAt   time.Time        `json:"created_at"`
}

type UserListItem struct {
	ID        int       `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	RoleID    int       `json:"role_id" db:"role_id"`
	RoleName  string    `json:"role_name" db:"role_name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CategoryListItem struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type UpdateUserRoleRequest struct {
	RoleID int `json:"role_id"`
}

type ResetPasswordRequest struct {
	NewPassword string `json:"new_password"`
}

type CreatePermissionRequest struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

type CreateCategoryRequest struct {
	Name string `json:"name"`
}

type UpdateCategoryRequest struct {
	Name string `json:"name"`
}
