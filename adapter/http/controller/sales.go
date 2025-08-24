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
