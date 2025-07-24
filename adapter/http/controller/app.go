package controller

import (
	"html/template"
	"net/http"
)

type appController struct {
	template *template.Template
}

func NewAppController(tmpl *template.Template) *appController {
	return &appController{template: tmpl}
}

func (c *appController) GetDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := c.template.ExecuteTemplate(w, "dashboard.html", nil)
	if err != nil {
		http.Error(w, "dashboard error", http.StatusInternalServerError)
	}
}
