package controller

import (
	"net/http"
	"pharmacy/adapter/http/helper"
	"pharmacy/httperror"
	"pharmacy/internal/types"
	"pharmacy/service"
	"strconv"
)

type customerController struct {
	service service.CustomerService
}

func NewCustomerController(svc service.CustomerService) *customerController {
	return &customerController{service: svc}
}

func customerID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		httperror.BadRequest("invalid customer id", err).JSONRespond(w)
		return 0, false
	}
	return id, true
}

func (c *customerController) ListCustomers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	includeInactive := query.Get("inactive") == "1" || query.Get("inactive") == "true"
	result, err := c.service.ListCustomers(r.Context(), query.Get("q"), includeInactive, page)
	if err != nil {
		respondServiceError(w, err, "could not fetch customers")
		return
	}
	helper.JSONResponse(w, http.StatusOK, result)
}

func (c *customerController) GetCustomer(w http.ResponseWriter, r *http.Request) {
	id, ok := customerID(w, r)
	if !ok {
		return
	}
	customer, err := c.service.GetCustomer(r.Context(), id)
	if err != nil {
		respondServiceError(w, err, "could not fetch customer")
		return
	}
	helper.JSONResponse(w, http.StatusOK, customer)
}

func (c *customerController) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var req types.CustomerRequest
	if !decodeLimitedJSON(w, r, &req) {
		return
	}
	id, err := c.service.CreateCustomer(r.Context(), req, requestUserID(r))
	if err != nil {
		respondServiceError(w, err, "could not save customer")
		return
	}
	helper.JSONResponse(w, http.StatusCreated, map[string]int{"id": id})
}

func (c *customerController) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	id, ok := customerID(w, r)
	if !ok {
		return
	}
	var req types.CustomerRequest
	if !decodeLimitedJSON(w, r, &req) {
		return
	}
	req.ID = id // the address says which customer; the body only carries the edit
	if err := c.service.UpdateCustomer(r.Context(), req, requestUserID(r)); err != nil {
		respondServiceError(w, err, "could not save customer")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *customerController) Deactivate(w http.ResponseWriter, r *http.Request) {
	c.setActive(w, r, false)
}

func (c *customerController) Reactivate(w http.ResponseWriter, r *http.Request) {
	c.setActive(w, r, true)
}

func (c *customerController) setActive(w http.ResponseWriter, r *http.Request, active bool) {
	id, ok := customerID(w, r)
	if !ok {
		return
	}
	var req types.CustomerActionRequest
	if !decodeLimitedJSON(w, r, &req) {
		return
	}
	err := c.service.SetCustomerActive(r.Context(), id, req.Version, active, requestUserID(r))
	if err != nil {
		respondServiceError(w, err, "could not save customer")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
