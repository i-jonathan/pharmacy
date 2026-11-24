package controller

import (
	"net/http"
	"pharmacy/adapter/http/helper"
	"pharmacy/httperror"
	"pharmacy/internal/types"
	"pharmacy/service"
	"strconv"
)

type healthCheckController struct {
	service service.HealthCheckService
}

func NewHealthCheckController(svc service.HealthCheckService) *healthCheckController {
	return &healthCheckController{service: svc}
}

// pathID reads the {id} path value, answering 400 itself when it is not a positive number.
func pathID(w http.ResponseWriter, r *http.Request, what string) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		httperror.BadRequest("invalid "+what+" id", err).JSONRespond(w)
		return 0, false
	}
	return id, true
}

func (c *healthCheckController) Record(w http.ResponseWriter, r *http.Request) {
	var req types.HealthCheckRequest
	if !decodeLimitedJSON(w, r, &req) {
		return
	}
	saved, err := c.service.RecordHealthCheck(r.Context(), req, requestUserID(r))
	if err != nil {
		respondServiceError(w, err, "could not save the health check")
		return
	}
	helper.JSONResponse(w, http.StatusCreated, saved)
}

func (c *healthCheckController) CustomerHistory(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "customer")
	if !ok {
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	result, err := c.service.CustomerHealthChecks(r.Context(), id, page, requestUserID(r))
	if err != nil {
		respondServiceError(w, err, "could not fetch health readings")
		return
	}
	helper.JSONResponse(w, http.StatusOK, result)
}

func (c *healthCheckController) Correct(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "health check")
	if !ok {
		return
	}
	var req types.HealthCheckCorrection
	if !decodeLimitedJSON(w, r, &req) {
		return
	}
	req.ID = id
	saved, err := c.service.CorrectHealthCheck(r.Context(), req, requestUserID(r))
	if err != nil {
		respondServiceError(w, err, "could not correct the health check")
		return
	}
	helper.JSONResponse(w, http.StatusOK, saved)
}

func (c *healthCheckController) Reprint(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "health check")
	if !ok {
		return
	}
	saved, err := c.service.ReprintHealthCheck(r.Context(), id, requestUserID(r))
	if err != nil {
		respondServiceError(w, err, "could not reprint the health check")
		return
	}
	helper.JSONResponse(w, http.StatusOK, saved)
}
