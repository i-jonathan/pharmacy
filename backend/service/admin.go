package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
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
	username, err := s.repo.FetchUsernameByID(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return httperror.NotFound("user not found", err)
	}
	if err != nil {
		return httperror.ServerError("failed to fetch user", err)
	}
	// Stored exactly as typed: login compares the untrimmed password.
	u := model.User{UserName: username, Password: newPassword}
	if err := u.ValidatePassword(); err != nil {
		return httperror.BadRequest(err.Error(), err)
	}

	if err := u.HashPassword(); err != nil {
		log.Println(err)
		return httperror.ServerError("failed to hash password", err)
	}

	err = s.repo.UpdateUserPassword(ctx, userID, u.Password)
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

// CreateUser adds a staff account with its role and starting password, and records who added
// it. actorID is 0 when there is no signed-in user (the create-admin command).
func (s *adminService) CreateUser(
	ctx context.Context, req types.CreateUserRequest, actorID int,
) (int, error) {
	user := model.User{UserName: strings.TrimSpace(req.Username), Password: req.Password}
	if err := user.ValidateUsername(); err != nil {
		return 0, httperror.BadRequest(err.Error(), err)
	}
	if err := user.ValidatePassword(); err != nil {
		return 0, httperror.BadRequest(err.Error(), err)
	}
	if req.RoleID <= 0 {
		return 0, httperror.BadRequest("choose a role", nil)
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return 0, httperror.ServerError("failed to start transaction", err)
	}
	committed := false
	defer func() {
		if !committed {
			s.repo.RollbackTx(tx)
		}
	}()

	if err = user.HashPassword(); err != nil {
		return 0, httperror.ServerError("failed to hash password", err)
	}
	user.RoleID = req.RoleID
	id, err := s.repo.CreateUserTx(ctx, tx, user)
	if errors.Is(err, repository.ErrUsernameTaken) {
		return 0, httperror.NewHttpError(http.StatusConflict, "username already exists", nil)
	}
	if errors.Is(err, repository.ErrRoleNotFound) {
		return 0, httperror.BadRequest("that role does not exist", err)
	}
	if err != nil {
		log.Println(err)
		return 0, httperror.ServerError("failed to create user", err)
	}

	details, err := json.Marshal(map[string]any{"username": user.UserName, "role_id": req.RoleID})
	if err != nil {
		return 0, httperror.ServerError("failed to record audit entry", err)
	}
	err = s.repo.InsertAuditLogTx(ctx, tx, types.AuditEntry{
		ActorID:    actorID,
		NoActor:    actorID == 0,
		EntityType: constant.AuditEntityUser,
		EntityID:   id,
		Action:     constant.AuditCreated,
		Details:    details,
	})
	if err != nil {
		return 0, httperror.ServerError("failed to record audit entry", err)
	}

	if err = s.repo.CommitTx(tx); err != nil {
		return 0, httperror.ServerError("failed to create user", err)
	}
	committed = true
	return id, nil
}
