package controller

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"pharmacy/adapter/http/helper"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/service"
	"strconv"
)

type adminController struct {
	adminService service.AdminService
	template     *template.Template
}

func NewAdminController(svc service.AdminService, tmpl *template.Template) *adminController {
	return &adminController{adminService: svc, template: tmpl}
}

func (c *adminController) GetAdminDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	perms, ok := r.Context().Value(constant.PermissionsSessionKey).(map[string]bool)
	if !ok {
		perms = make(map[string]bool)
	}

	data := struct {
		Permissions map[string]bool
		Title       string
		ActivePage  string
	}{
		Permissions: perms,
		Title:       "Admin Dashboard",
		ActivePage:  "admin",
	}

	err := c.template.ExecuteTemplate(w, "admin.html", data)
	if err != nil {
		http.Error(w, "admin page render error", http.StatusInternalServerError)
	}
}

func (c *adminController) ListModules(w http.ResponseWriter, r *http.Request) {
	modules := c.adminService.GetModules(r.Context())
	helper.JSONResponse(w, http.StatusOK, modules)
}

func (c *adminController) ListPermissions(w http.ResponseWriter, r *http.Request) {
	permissions, err := c.adminService.GetPermissions(r.Context())
	if err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			httperr.JSONRespond(w)
			return
		}
		httperror.ServerError("failed to fetch permissions", err).JSONRespond(w)
		return
	}
	helper.JSONResponse(w, http.StatusOK, permissions)
}

func (c *adminController) CreatePermission(w http.ResponseWriter, r *http.Request) {
	var params types.CreatePermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httperror.BadRequest("invalid json", err).JSONRespond(w)
		return
	}

	if err := c.adminService.CreatePermission(r.Context(), params.Resource, params.Action); err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			httperr.JSONRespond(w)
			return
		}
		httperror.ServerError("failed to create permission", err).JSONRespond(w)
		return
	}

	helper.JSONResponse(w, http.StatusOK, map[string]string{"message": "Permission created"})
}

func (c *adminController) DeletePermission(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httperror.BadRequest("invalid permission id", err).JSONRespond(w)
		return
	}

	if err := c.adminService.DeletePermission(r.Context(), id); err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			httperr.JSONRespond(w)
			return
		}
		httperror.ServerError("failed to delete permission", err).JSONRespond(w)
		return
	}

	helper.JSONResponse(w, http.StatusOK, map[string]string{"message": "Permission deleted"})
}

func (c *adminController) AssignPermissionToRole(w http.ResponseWriter, r *http.Request) {
	permID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httperror.BadRequest("invalid permission id", err).JSONRespond(w)
		return
	}
	roleID, err := strconv.Atoi(r.PathValue("roleId"))
	if err != nil {
		httperror.BadRequest("invalid role id", err).JSONRespond(w)
		return
	}

	if err := c.adminService.AssignPermissionToRole(r.Context(), roleID, permID); err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			httperr.JSONRespond(w)
			return
		}
		httperror.ServerError("failed to assign permission", err).JSONRespond(w)
		return
	}

	helper.JSONResponse(w, http.StatusOK, map[string]string{"message": "Permission assigned"})
}

func (c *adminController) RemovePermissionFromRole(w http.ResponseWriter, r *http.Request) {
	permID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httperror.BadRequest("invalid permission id", err).JSONRespond(w)
		return
	}
	roleID, err := strconv.Atoi(r.PathValue("roleId"))
	if err != nil {
		httperror.BadRequest("invalid role id", err).JSONRespond(w)
		return
	}

	if err := c.adminService.RemovePermissionFromRole(r.Context(), roleID, permID); err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			httperr.JSONRespond(w)
			return
		}
		httperror.ServerError("failed to remove permission", err).JSONRespond(w)
		return
	}

	helper.JSONResponse(w, http.StatusOK, map[string]string{"message": "Permission removed"})
}

func (c *adminController) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := c.adminService.GetRoles(r.Context())
	if err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			httperr.JSONRespond(w)
			return
		}
		httperror.ServerError("failed to fetch roles", err).JSONRespond(w)
		return
	}
	helper.JSONResponse(w, http.StatusOK, roles)
}

func (c *adminController) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := c.adminService.GetUsers(r.Context())
	if err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			httperr.JSONRespond(w)
			return
		}
		httperror.ServerError("failed to fetch users", err).JSONRespond(w)
		return
	}
	helper.JSONResponse(w, http.StatusOK, users)
}

func (c *adminController) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httperror.BadRequest("invalid user id", err).JSONRespond(w)
		return
	}

	var params types.UpdateUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httperror.BadRequest("invalid json", err).JSONRespond(w)
		return
	}

	if err := c.adminService.UpdateUserRole(r.Context(), userID, params.RoleID); err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			httperr.JSONRespond(w)
			return
		}
		httperror.ServerError("failed to update user role", err).JSONRespond(w)
		return
	}

	helper.JSONResponse(w, http.StatusOK, map[string]string{"message": "User role updated"})
}

func (c *adminController) ResetUserPassword(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httperror.BadRequest("invalid user id", err).JSONRespond(w)
		return
	}

	var params types.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httperror.BadRequest("invalid json", err).JSONRespond(w)
		return
	}

	if err := c.adminService.ResetUserPassword(r.Context(), userID, params.NewPassword); err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			httperr.JSONRespond(w)
			return
		}
		httperror.ServerError("failed to reset password", err).JSONRespond(w)
		return
	}

	helper.JSONResponse(w, http.StatusOK, map[string]string{"message": "Password reset successfully"})
}

func (c *adminController) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := c.adminService.GetCategories(r.Context())
	if err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			httperr.JSONRespond(w)
			return
		}
		httperror.ServerError("failed to fetch categories", err).JSONRespond(w)
		return
	}
	helper.JSONResponse(w, http.StatusOK, categories)
}

func (c *adminController) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var params types.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httperror.BadRequest("invalid json", err).JSONRespond(w)
		return
	}

	if err := c.adminService.CreateCategory(r.Context(), params.Name); err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			httperr.JSONRespond(w)
			return
		}
		httperror.ServerError("failed to create category", err).JSONRespond(w)
		return
	}

	helper.JSONResponse(w, http.StatusOK, map[string]string{"message": "Category created"})
}

func (c *adminController) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httperror.BadRequest("invalid category id", err).JSONRespond(w)
		return
	}

	var params types.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httperror.BadRequest("invalid json", err).JSONRespond(w)
		return
	}

	if err := c.adminService.UpdateCategory(r.Context(), id, params.Name); err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			httperr.JSONRespond(w)
			return
		}
		httperror.ServerError("failed to update category", err).JSONRespond(w)
		return
	}

	helper.JSONResponse(w, http.StatusOK, map[string]string{"message": "Category updated"})
}

func (c *adminController) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httperror.BadRequest("invalid category id", err).JSONRespond(w)
		return
	}

	if err := c.adminService.DeleteCategory(r.Context(), id); err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			httperr.JSONRespond(w)
			return
		}
		httperror.ServerError("failed to delete category", err).JSONRespond(w)
		return
	}

	helper.JSONResponse(w, http.StatusOK, map[string]string{"message": "Category deleted"})
}
