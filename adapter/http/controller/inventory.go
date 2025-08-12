package controller

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"pharmacy/adapter/http/helper"
	"pharmacy/httperror"
	"pharmacy/internal/types"
	"pharmacy/service"
)

type inventoryController struct {
	service  service.InventoryService
	template *template.Template
}

func NewInventoryController(svc service.InventoryService, tmpl *template.Template) *inventoryController {
	return &inventoryController{service: svc, template: tmpl}
}

func (c *inventoryController) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var params types.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httperror.BadRequest("invalid json", err).JSONRespond(w)
		return
	}

	itemResponse, err := c.service.CreateProduct(r.Context(), params)
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

	helper.JSONResponse(w, http.StatusOK, itemResponse)
}
