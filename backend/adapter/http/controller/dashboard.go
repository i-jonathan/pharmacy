package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"pharmacy/httperror"
	"pharmacy/service"
)

type dashboardController struct {
	dashboardService service.DashboardService
}

func NewDashboardController(dashboardService service.DashboardService) *dashboardController {
	return &dashboardController{dashboardService: dashboardService}
}

func (c *dashboardController) GetDashboardData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := c.dashboardService.GetDashboardData(ctx)
	if err != nil {
		var httperr *httperror.HTTPError
		if errors.As(err, &httperr) {
			httperr.JSONRespond(w)
			return
		}

		httperror.ServerError("failed to get dashboard data", err).JSONRespond(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}
