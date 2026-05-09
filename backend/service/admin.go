package service

import (
	"context"
	"log"
	"pharmacy/httperror"
	"pharmacy/internal/types"
	"pharmacy/model"
	"pharmacy/repository"
	"strings"
)

type adminService struct {
	repo repository.PharmacyRepository
}

func NewAdminService(repo repository.PharmacyRepository) AdminService {
	return &adminService{repo: repo}
}

func (s *adminService) GetModules(_ context.Context) []types.AdminModuleInfo {
	return []types.AdminModuleInfo{
		{Name: "Permissions", Description: "Manage access permissions and assign them to roles", Icon: "shield", Path: "permissions"},
		{Name: "Roles", Description: "View roles and their permission assignments", Icon: "users", Path: "roles"},
		{Name: "Users", Description: "Manage users, change roles, reset passwords", Icon: "user-cog", Path: "users"},
		{Name: "Categories", Description: "Manage product categories", Icon: "tags", Path: "categories"},
	}
}

func (s *adminService) GetPermissions(ctx context.Context) ([]types.PermissionInfo, error) {
	perms, err := s.repo.GetAllPermissions(ctx)
	if err != nil {
		log.Println(err)
		return nil, httperror.ServerError("failed to fetch permissions", err)
	}

	result := make([]types.PermissionInfo, 0, len(perms))
	for _, p := range perms {
		result = append(result, types.PermissionInfo{
			ID:       p.ID,
			Resource: p.Resource,
			Action:   p.Action,
		})
	}
	return result, nil
}

func (s *adminService) CreatePermission(ctx context.Context, resource, action string) error {
	resource = strings.TrimSpace(resource)
	action = strings.TrimSpace(action)

	if resource == "" || action == "" {
		return httperror.BadRequest("resource and action are required", nil)
	}

	_, err := s.repo.CreatePermission(ctx, resource, action)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("failed to create permission", err)
	}
	return nil
}

func (s *adminService) DeletePermission(ctx context.Context, id int) error {
	err := s.repo.DeletePermission(ctx, id)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("failed to delete permission", err)
	}
	return nil
}

func (s *adminService) GetRolePermissions(ctx context.Context, roleID int) ([]types.PermissionInfo, error) {
	perms, err := s.repo.GetRolePermissions(ctx, roleID)
	if err != nil {
		log.Println(err)
		return nil, httperror.ServerError("failed to fetch role permissions", err)
	}

	result := make([]types.PermissionInfo, 0, len(perms))
	for _, p := range perms {
		result = append(result, types.PermissionInfo{
			ID:       p.ID,
			Resource: p.Resource,
			Action:   p.Action,
		})
	}
	return result, nil
}

func (s *adminService) AssignPermissionToRole(ctx context.Context, roleID, permissionID int) error {
	err := s.repo.AssignPermissionToRole(ctx, roleID, permissionID)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("failed to assign permission", err)
	}
	return nil
}

func (s *adminService) RemovePermissionFromRole(ctx context.Context, roleID, permissionID int) error {
	err := s.repo.RemovePermissionFromRole(ctx, roleID, permissionID)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("failed to remove permission", err)
	}
	return nil
}

func (s *adminService) GetRoles(ctx context.Context) ([]types.RoleWithPermissions, error) {
	roles, err := s.repo.GetAllRolesWithPermissions(ctx)
	if err != nil {
		log.Println(err)
		return nil, httperror.ServerError("failed to fetch roles", err)
	}
	return roles, nil
}

func (s *adminService) GetUsers(ctx context.Context) ([]types.UserListItem, error) {
	users, err := s.repo.ListUsers(ctx)
	if err != nil {
		log.Println(err)
		return nil, httperror.ServerError("failed to fetch users", err)
	}
	return users, nil
}

func (s *adminService) UpdateUserRole(ctx context.Context, userID, roleID int) error {
	err := s.repo.UpdateUserRole(ctx, userID, roleID)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("failed to update user role", err)
	}
	return nil
}

func (s *adminService) ResetUserPassword(ctx context.Context, userID int, newPassword string) error {
	newPassword = strings.TrimSpace(newPassword)
	if len(newPassword) < 8 {
		return httperror.BadRequest("password must be at least 8 characters", nil)
	}

	u := model.User{Password: newPassword}
	if err := u.HashPassword(); err != nil {
		log.Println(err)
		return httperror.ServerError("failed to hash password", err)
	}

	err := s.repo.UpdateUserPassword(ctx, userID, u.Password)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("failed to update password", err)
	}
	return nil
}

func (s *adminService) GetCategories(ctx context.Context) ([]types.CategoryListItem, error) {
	cats, err := s.repo.FetchProductCategories(ctx)
	if err != nil {
		log.Println(err)
		return nil, httperror.ServerError("failed to fetch categories", err)
	}

	result := make([]types.CategoryListItem, 0, len(cats))
	for _, c := range cats {
		result = append(result, types.CategoryListItem{
			ID:        c.ID,
			Name:      c.Name,
			CreatedAt: c.CreatedAt,
		})
	}
	return result, nil
}

func (s *adminService) CreateCategory(ctx context.Context, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return httperror.BadRequest("category name is required", nil)
	}

	_, err := s.repo.CreateCategory(ctx, name)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("failed to create category", err)
	}
	return nil
}

func (s *adminService) UpdateCategory(ctx context.Context, id int, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return httperror.BadRequest("category name is required", nil)
	}

	err := s.repo.UpdateCategory(ctx, id, name)
	if err != nil {
		log.Println(err)
		return httperror.ServerError("failed to update category", err)
	}
	return nil
}

func (s *adminService) DeleteCategory(ctx context.Context, id int) error {
	err := s.repo.DeleteCategory(ctx, id)
	if err != nil {
		log.Println(err)
		errMsg := err.Error()
		if strings.Contains(errMsg, "foreign key") || strings.Contains(errMsg, "constraint") {
			return httperror.BadRequest("cannot delete category: it is used by existing products", err)
		}
		return httperror.ServerError("failed to delete category", err)
	}
	return nil
}
