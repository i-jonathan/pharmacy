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

func (c *inventoryController) GetReceiveItems(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	categories, err := c.service.FetchCategories(r.Context())
	if err != nil {
		http.Error(w, "error fetching categories", http.StatusInternalServerError)
		return
	}

	data := struct {
		Categories []types.CategoriesResponse
	}{
		Categories: categories,
	}

	err = c.template.ExecuteTemplate(w, "receive-items.html", data)
	if err != nil {
		http.Error(w, "receive items render error", http.StatusInternalServerError)
	}
}

func (c *inventoryController) SearchForProduct(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		httperror.BadRequest("Missing query string", nil).JSONRespond(w)
		return
	}

	products, err := c.service.SearchProducts(r.Context(), query)
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

	helper.JSONResponse(w, http.StatusOK, products)
}
