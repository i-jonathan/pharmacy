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
	"time"
)

type saleController struct {
	service  service.SaleService
	template *template.Template
}

func NewSaleController(svc service.SaleService, tmpl *template.Template) *saleController {
	return &saleController{svc, tmpl}
}

func (c *saleController) RenderSalesReceipt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := c.template.ExecuteTemplate(w, "receipt.html", nil)
	if err != nil {
		http.Error(w, "sales receipt render error", http.StatusInternalServerError)
	}
}

func (c *saleController) CreateSale(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(constant.UserIDKey)
	uID, ok := userID.(int)
	if !ok {
		httperror.Unauthorized("User ID is missing in context", nil).JSONRespond(w)
		return
	}

	var saleParams types.Sale

	if err := json.NewDecoder(r.Body).Decode(&saleParams); err != nil {
		httperror.BadRequest("Invalid JSON", err).JSONRespond(w)
		return
	}

	saleParams.CashierID = uID

	err := c.service.CreateSale(r.Context(), saleParams)
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

	helper.JSONResponse(w, http.StatusOK, map[string]any{"msg": "Saved Successfully"})
}

func (c *saleController) RenderSalesHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	sales, err := c.service.FetchSalesHistory(r.Context(), types.SaleFilter{})
	if err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			http.Error(w, httperr.Message, httperr.Code)
			return
		}

		http.Error(w, "sales history fetch error", http.StatusInternalServerError)
		return
	}

	salesJSON, err := json.Marshal(sales)
	if err != nil {
		http.Error(w, "sales history json error", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"SalesJSON": template.JS(salesJSON),
	}

	err = c.template.ExecuteTemplate(w, "sales-history.html", data)
	if err != nil {
		http.Error(w, "sales history render error", http.StatusInternalServerError)
	}
}

func (c *saleController) FilterSales(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	var filter types.SaleFilter

	if startStr != "" {
		if start, err := time.Parse("2006-01-02", startStr); err == nil {
			filter.StartDate = &start
		}
	}
	if endStr != "" {
		if end, err := time.Parse("2006-01-02", endStr); err == nil {
			filter.EndDate = &end
		}
	}

	salesData, err := c.service.FetchSalesHistory(r.Context(), filter)
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

	helper.JSONResponse(w, http.StatusOK, salesData)
}

func (c *saleController) HoldSaleTransaction(w http.ResponseWriter, r *http.Request) {
	var heldTransaction types.HoldTransactionRequest

	if err := json.NewDecoder(r.Body).Decode(&heldTransaction); err != nil {
		httperror.BadRequest("Invalid JSON", err).JSONRespond(w)
		return
	}

	err := c.service.HoldSale(r.Context(), heldTransaction)
	if err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			http.Error(w, httperr.Message, httperr.Code)
			return
		}

		http.Error(w, "sales history fetch error", http.StatusInternalServerError)
		return
	}

	helper.JSONResponse(w, http.StatusOK, map[string]string{"message": "Held Successfully"})
}

func (c *saleController) RenderHeldSaleReceipts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	heldTransactions, err := c.service.FetchHeldSaleTransactions(r.Context())
	if err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			http.Error(w, httperr.Message, httperr.Code)
			return
		}

		http.Error(w, "sales history fetch error", http.StatusInternalServerError)
		return
	}

	transactionJSON, err := json.Marshal(heldTransactions)
	if err != nil {
		http.Error(w, "held transactions json error", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"HeldTransactions": template.JS(transactionJSON),
	}

	err = c.template.ExecuteTemplate(w, "held-receipt.html", data)
	if err != nil {
		http.Error(w, "held receipts render error", http.StatusInternalServerError)
	}
}

func (c *saleController) DeleteHeldSale(w http.ResponseWriter, r *http.Request) {
	reference := r.PathValue("reference")
	if reference == "" {
		http.Error(w, "missing reference", http.StatusBadRequest)
		return
	}

	err := c.service.DeleteHeldTransaction(r.Context(), reference)
	if err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			http.Error(w, httperr.Message, httperr.Code)
			return
		}

		http.Error(w, "sales history fetch error", http.StatusInternalServerError)
		return
	}
}

func (c *saleController) ReturnItems(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(constant.UserIDKey)
	uID, ok := userID.(int)
	if !ok {
		httperror.Unauthorized("User ID is missing in context", nil).JSONRespond(w)
		return
	}

	var returnParams types.ReturnSale

	if err := json.NewDecoder(r.Body).Decode(&returnParams); err != nil {
		httperror.BadRequest("Invalid JSON", err).JSONRespond(w)
		return
	}

	returnParams.CashierID = uID

	err := c.service.ReturnItems(r.Context(), returnParams)
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

	helper.JSONResponse(w, http.StatusOK, map[string]any{"msg": "Items Successfully returned"})
}
