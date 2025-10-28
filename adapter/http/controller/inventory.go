package controller

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"pharmacy/adapter/http/helper"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/internal/types"
	"pharmacy/model"
	"pharmacy/service"
	"strconv"
	"time"
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

func (c *inventoryController) SearchForSuppliers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		httperror.BadRequest("Missing search query in parameters", nil).JSONRespond(w)
		return
	}

	suppliers, err := c.service.SearchForSuppliers(r.Context(), query)
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

	helper.JSONResponse(w, http.StatusOK, map[string]any{
		"data": suppliers,
	})
}

func (c *inventoryController) ReceiveSupply(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(constant.UserIDKey)
	uid, ok := userID.(int)
	if !ok {
		httperror.Unauthorized("User ID is missing in context", nil).JSONRespond(w)
		return
	}

	var params types.ReceiveSupplyRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httperror.BadRequest("invalid json", err).JSONRespond(w)
		return
	}
	params.UserID = uid
	err := c.service.ReceiveProductSupply(r.Context(), params)
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

	helper.JSONResponse(w, http.StatusOK, map[string]string{"message": "Items saved successfully"})
}

func (c *inventoryController) RenderInventoryPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	inventory, err := c.service.FetchInventory(r.Context())
	if err != nil {
		http.Error(w, "error fetching inventory", http.StatusInternalServerError)
		return
	}

	data := struct {
		Categories []model.Category
		Items      []model.InventoryItem
	}{
		Categories: inventory.Categories,
		Items:      inventory.Items,
	}

	err = c.template.ExecuteTemplate(w, "inventory.html", data)
	if err != nil {
		http.Error(w, "inventory page render error", http.StatusInternalServerError)
	}
}

func (c *inventoryController) DownloadInventoryReport(w http.ResponseWriter, r *http.Request) {
	inventory, err := c.service.FetchInventory(r.Context())
	if err != nil {
		http.Error(w, "error fetching inventory", http.StatusInternalServerError)
		return
	}

	now := time.Now()
	filename := fmt.Sprintf("attachment; filename=Stock-%s.csv", now.Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", filename)

	writer := csv.NewWriter(w)
	defer writer.Flush()

	writer.Write([]string{"ID", "Category", "Name", "Price", "Qty.", "Expiry Date", "Store Qty.", "Disp. Qty"})
	for _, item := range inventory.Items {
		expiry := ""
		if item.EarliestExpiry != nil {
			expiry = item.EarliestExpiry.Format("01/2006")
		}

		writer.Write([]string{
			strconv.Itoa(item.ID),
			item.Category,
			item.Name,
			strconv.FormatFloat(float64(item.DefaultPrice)/100, 'f', 2, 64),
			strconv.Itoa(item.Stock),
			expiry,
		})
	}
}
