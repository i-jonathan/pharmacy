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
	"pharmacy/model"
	"pharmacy/service"
	"strconv"
	"strings"
)

type stockTakingController struct {
	service  service.StockTakingService
	template *template.Template
}

func NewStockTakingController(service service.StockTakingService, tmpl *template.Template) *stockTakingController {
	return &stockTakingController{
		service:  service,
		template: tmpl,
	}
}

func (c *stockTakingController) RenderStockTakingDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]any{
		"Title":         "Stock Taking Dashboard",
		"ActivePage":    "stock-taking",
		"SubActivePage": "dashboard",
	}
	err := c.template.ExecuteTemplate(w, "stock-taking-dashboard.html", data)
	if err != nil {
		http.Error(w, "stock taking dashboard render error", http.StatusInternalServerError)
	}
}

func (c *stockTakingController) ListStockTakings(w http.ResponseWriter, r *http.Request) {
	items, err := c.service.ListAllStockTakings(r.Context())
	if err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			httperr.JSONRespond(w)
			return
		}

		httperror.ServerError("fetching stock takings failed", err).JSONRespond(w)
		return
	}

	helper.JSONResponse(w, http.StatusOK, map[string]any{
		"stock_takings": items,
	})
}

func (c *stockTakingController) CreateStockTaking(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(constant.UserIDKey)
	uid, ok := userID.(int)
	if !ok {
		httperror.Unauthorized("User ID is missing in context", nil).JSONRespond(w)
		return
	}

	var req types.CreateStockTakingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperror.BadRequest("invalid json", err).JSONRespond(w)
		return
	}

	data := types.StockTakingData{
		Name:        req.Name,
		CreatedByID: uid,
	}

	id, err := c.service.CreateStockTaking(r.Context(), data)
	if err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			httperr.JSONRespond(w)
			return
		}

		httperror.ServerError("failed to create stock taking", err).JSONRespond(w)
		return
	}

	helper.JSONResponse(w, http.StatusCreated, map[string]any{
		"id": id,
	})
}

func (c *stockTakingController) FetchStockTaking(w http.ResponseWriter, r *http.Request) {
	stockTakingID := r.PathValue("id")
	if strings.TrimSpace(stockTakingID) == "" {
		httperror.BadRequest("invalid stock taking id provided", nil).JSONRespond(w)
		return
	}

	id, err := strconv.Atoi(stockTakingID)
	if err != nil {
		httperror.BadRequest("invalid stock taking id provided", err).JSONRespond(w)
		return
	}

	data, err := c.service.FetchStockTaking(r.Context(), id)
	if err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			httperr.JSONRespond(w)
			return
		}

		httperror.ServerError("fetching stock taking data failed", err).JSONRespond(w)
		return
	}

	items, err := c.service.FetchStockTakingItems(r.Context(), id)
	if err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			httperr.JSONRespond(w)
			return
		}

		httperror.ServerError("fetching stock taking items failed", err).JSONRespond(w)
		return
	}

	perms, _ := r.Context().Value(constant.PermissionsSessionKey).(map[string]bool)

	helper.JSONResponse(w, http.StatusOK, types.StockTakingResponse{
		StockTakingData: data,
		Items:           items,
		Permissions:     perms,
	})
}

func (c *stockTakingController) FetchStockTakingItems(w http.ResponseWriter, r *http.Request) {
	stockTakingID := r.PathValue("id")
	if strings.TrimSpace(stockTakingID) == "" {
		httperror.BadRequest("invalid stock taking id provided", nil).JSONRespond(w)
		return
	}

	id, err := strconv.Atoi(stockTakingID)
	if err != nil {
		httperror.BadRequest("invalid stock taking id provided", err).JSONRespond(w)
		return
	}

	items, err := c.service.FetchStockTakingItems(r.Context(), id)
	if err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			httperr.JSONRespond(w)
			return
		}

		httperror.ServerError("fetching stock taking items failed", err).JSONRespond(w)
		return
	}

	helper.JSONResponse(w, http.StatusOK, map[string]any{
		"items": items,
	})
}

func (c *stockTakingController) RenderStockTakingPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	id := r.PathValue("id")

	data := struct {
		StockTakingID string
		Title         string
		ActivePage    string
		SubActivePage string
	}{
		StockTakingID: id,
		Title:         "Stock Taking",
		ActivePage:    "stock-taking",
		SubActivePage: "stock-taking-page",
	}

	err := c.template.ExecuteTemplate(w, "stock-taking.html", data)
	if err != nil {
		http.Error(w, "stock taking page render error", http.StatusInternalServerError)
	}
}

func (c *stockTakingController) UpdateStockTakingItemCount(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(constant.UserIDKey)
	uid, ok := userID.(int)
	if !ok {
		httperror.Unauthorized("User ID is missing in context", nil).JSONRespond(w)
		return
	}

	stockTakingID := r.PathValue("id")
	if strings.TrimSpace(stockTakingID) == "" {
		httperror.BadRequest("invalid stock taking id provided", nil).JSONRespond(w)
		return
	}

	id, err := strconv.Atoi(stockTakingID)
	if err != nil {
		httperror.BadRequest("invalid stock taking id provided", err).JSONRespond(w)
		return
	}

	productIDStr := r.PathValue("product_id")
	if strings.TrimSpace(productIDStr) == "" {
		httperror.BadRequest("invalid product id provided", nil).JSONRespond(w)
		return
	}

	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		httperror.BadRequest("invalid product id provided", err).JSONRespond(w)
		return
	}

	var data types.StockTakingItemCount
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		httperror.BadRequest("invalid json", err).JSONRespond(w)
		return
	}

	data.ProductID = productID
	data.StockTakingID = id
	data.UpdatedByID = uid

	err = c.service.UpdateStockTakingItemCount(r.Context(), data)
	if err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			httperr.JSONRespond(w)
			return
		}

		httperror.ServerError("failed to update stock taking item count", err)
		return
	}

	helper.JSONResponse(w, http.StatusOK, map[string]any{
		"status": "Updated",
	})
}

func (c *stockTakingController) CompleteStockTaking(w http.ResponseWriter, r *http.Request) {
	uid := r.Context().Value(constant.UserIDKey)
	userID, ok := uid.(int)
	if !ok {
		httperror.Unauthorized("User ID is missing in context", nil).JSONRespond(w)
		return
	}

	stockTakingID := r.PathValue("id")
	if strings.TrimSpace(stockTakingID) == "" {
		httperror.BadRequest("invalid stock taking id provided", nil).JSONRespond(w)
		return
	}

	stockID, err := strconv.Atoi(stockTakingID)
	if err != nil {
		httperror.BadRequest("invalid stock taking id provided", err).JSONRespond(w)
		return
	}

	err = c.service.CompleteStockTaking(r.Context(), stockID, userID)
	if err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			httperr.JSONRespond(w)
			return
		}

		httperror.ServerError("failed to update stock taking item count", err)
		return
	}

	helper.JSONResponse(w, http.StatusOK, map[string]any{
		"status": model.StockTakingCompleted,
	})
}
