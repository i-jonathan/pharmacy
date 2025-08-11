package controller

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"pharmacy/httperror"
	"pharmacy/internal/types"
	"pharmacy/service"
)

type inventoryController struct {
	service  service.InventoryService
	template *template.Template
}

func NewInventoryController(tmpl *template.Template) *inventoryController {
	return &inventoryController{template: tmpl}
}

func (c *inventoryController) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var params types.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httperror.BadRequest("invalid json", err).JSONRespond(w)
		return
	}

	err := c.service.CreateProduct(r.Context(), params)
	if err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			httperr.JSONRespond(w)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{"error": err})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
