package controller

import (
	"net/http"
	"pharmacy/adapter/http/helper"
	"pharmacy/httperror"
	"pharmacy/internal/types"
	"strconv"
)

// maxStockAdjustmentBody fits an adjustment of 200 products with room to spare.
const maxStockAdjustmentBody = 128 * 1024

func (c *inventoryController) CreateStockAdjustment(w http.ResponseWriter, r *http.Request) {
	var req types.StockAdjustmentRequest
	if !decodeJSONWithin(w, r, &req, maxStockAdjustmentBody) {
		return
	}
	id, err := c.service.CreateStockAdjustment(r.Context(), req, requestUserID(r))
	if err != nil {
		respondServiceError(w, err, "could not save stock adjustment")
		return
	}
	helper.JSONResponse(w, http.StatusCreated, map[string]int{"id": id})
}

func (c *inventoryController) ListStockAdjustments(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	result, err := c.service.ListStockAdjustments(r.Context(), page)
	if err != nil {
		respondServiceError(w, err, "could not fetch stock adjustments")
		return
	}
	helper.JSONResponse(w, http.StatusOK, result)
}

func (c *inventoryController) FlagStock(w http.ResponseWriter, r *http.Request) {
	var req types.StockFlagRequest
	if !decodeLimitedJSON(w, r, &req) {
		return
	}
	id, err := c.service.FlagStock(r.Context(), req, requestUserID(r))
	if err != nil {
		respondServiceError(w, err, "could not flag stock")
		return
	}
	helper.JSONResponse(w, http.StatusCreated, map[string]int{"id": id})
}

func (c *inventoryController) ListStockFlags(w http.ResponseWriter, r *http.Request) {
	flags, err := c.service.ListStockFlags(r.Context())
	if err != nil {
		respondServiceError(w, err, "could not fetch flags")
		return
	}
	helper.JSONResponse(w, http.StatusOK, flags)
}

func (c *inventoryController) CountStockFlags(w http.ResponseWriter, r *http.Request) {
	n, err := c.service.CountStockFlags(r.Context())
	if err != nil {
		respondServiceError(w, err, "could not count flags")
		return
	}
	helper.JSONResponse(w, http.StatusOK, map[string]int{"open": n})
}

func (c *inventoryController) DismissStockFlag(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httperror.BadRequest("invalid flag id", err).JSONRespond(w)
		return
	}
	var req types.DismissStockFlagRequest
	if !decodeLimitedJSON(w, r, &req) {
		return
	}
	if err = c.service.DismissStockFlag(r.Context(), id, req.Note, requestUserID(r)); err != nil {
		respondServiceError(w, err, "could not dismiss flag")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
