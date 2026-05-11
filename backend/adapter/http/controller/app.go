package controller

import (
	"encoding/json"
	"html/template"
	"net/http"
	"pharmacy/config"
	"pharmacy/internal/constant"
)

type appController struct {
	template *template.Template
}

func NewAppController(tmpl *template.Template) *appController {
	return &appController{template: tmpl}
}

func (c *appController) GetDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	perms := getPermissionsFromContext(r)
	userID := getUserIDFromContext(r)

	store := config.NewSessionStore()
	session, _ := store.Get(r, "session")
	userName, _ := session.Values[constant.UserNameSessionKey].(string)
	roleName, _ := session.Values[constant.RoleNameSessionKey].(string)

	data := map[string]any{
		"Title":       "Dashboard",
		"ActivePage":  "dashboard",
		"Permissions": perms,
		"User": map[string]any{
			"id":       userID,
			"username": userName,
			"role":     roleName,
		},
	}

	ui := r.URL.Query().Get("ui")
	if ui == "v2" {
		err := c.template.ExecuteTemplate(w, "next-dashboard.html", data)
		if err != nil {
			http.Error(w, "dashboard error", http.StatusInternalServerError)
		}
		return
	}

	err := c.template.ExecuteTemplate(w, "dashboard.html", data)
	if err != nil {
		http.Error(w, "dashboard error", http.StatusInternalServerError)
	}
}

func getPermissionsFromContext(r *http.Request) map[string]bool {
	perms, ok := r.Context().Value(constant.PermissionsSessionKey).(map[string]bool)
	if !ok || perms == nil {
		return make(map[string]bool)
	}
	return perms
}

func getUserIDFromContext(r *http.Request) int {
	userID, ok := r.Context().Value(constant.UserIDKey).(int)
	if !ok {
		return 0
	}
	return userID
}

func GetPermissionsAsJSON(r *http.Request) string {
	perms := getPermissionsFromContext(r)
	b, err := json.Marshal(perms)
	if err != nil {
		return "{}"
	}
	return string(b)
}
