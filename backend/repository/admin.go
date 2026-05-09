package repository

import (
	"context"
	"encoding/json"
	"pharmacy/internal/types"
	"pharmacy/model"
	"time"
)

func (r *repo) GetAllPermissions(ctx context.Context) ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.Data.SelectContext(ctx, &permissions, getAllPermissionsQuery)
	return permissions, err
}

func (r *repo) CreatePermission(ctx context.Context, resource, action string) (int, error) {
	var id int
	err := r.Data.GetContext(ctx, &id, createPermissionQuery, resource, action)
	return id, err
}

func (r *repo) DeletePermission(ctx context.Context, id int) error {
	_, err := r.Data.ExecContext(ctx, deletePermissionQuery, id)
	return err
}

func (r *repo) GetRolePermissions(ctx context.Context, roleID int) ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.Data.SelectContext(ctx, &permissions, getRolePermissionsQuery, roleID)
	return permissions, err
}

func (r *repo) AssignPermissionToRole(ctx context.Context, roleID, permissionID int) error {
	_, err := r.Data.ExecContext(ctx, assignPermissionToRoleQuery, roleID, permissionID)
	return err
}

func (r *repo) RemovePermissionFromRole(ctx context.Context, roleID, permissionID int) error {
	_, err := r.Data.ExecContext(ctx, removePermissionFromRoleQuery, roleID, permissionID)
	return err
}

func (r *repo) GetAllRolesWithPermissions(ctx context.Context) ([]types.RoleWithPermissions, error) {
	type row struct {
		ID          int             `db:"id"`
		Name        string          `db:"name"`
		CreatedAt   time.Time       `db:"created_at"`
		Permissions json.RawMessage `db:"permissions"`
	}

	var rows []row
	err := r.Data.SelectContext(ctx, &rows, getAllRolesWithPermissionsQuery)
	if err != nil {
		return nil, err
	}

	result := make([]types.RoleWithPermissions, 0, len(rows))
	for _, rw := range rows {
		var permissions []types.PermissionInfo
		if err := json.Unmarshal(rw.Permissions, &permissions); err != nil {
			return nil, err
		}

		result = append(result, types.RoleWithPermissions{
			ID:          rw.ID,
			Name:        rw.Name,
			Permissions: permissions,
			CreatedAt:   rw.CreatedAt,
		})
	}
	return result, nil
}

func (r *repo) ListUsers(ctx context.Context) ([]types.UserListItem, error) {
	var users []types.UserListItem
	err := r.Data.SelectContext(ctx, &users, listUsersQuery)
	return users, err
}

func (r *repo) UpdateUserRole(ctx context.Context, userID, roleID int) error {
	_, err := r.Data.ExecContext(ctx, updateUserRoleQuery, roleID, userID)
	return err
}

func (r *repo) UpdateUserPassword(ctx context.Context, userID int, hashedPassword string) error {
	_, err := r.Data.ExecContext(ctx, updateUserPasswordQuery, hashedPassword, userID)
	return err
}

func (r *repo) CreateCategory(ctx context.Context, name string) (int, error) {
	var id int
	err := r.Data.GetContext(ctx, &id, createCategoryQuery, name)
	return id, err
}

func (r *repo) UpdateCategory(ctx context.Context, id int, name string) error {
	_, err := r.Data.ExecContext(ctx, updateCategoryQuery, name, id)
	return err
}

func (r *repo) DeleteCategory(ctx context.Context, id int) error {
	_, err := r.Data.ExecContext(ctx, deleteCategoryQuery, id)
	return err
}
