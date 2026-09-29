package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"pharmacy/service"
	"time"
)

type dashboardController struct {
	dashboardService service.DashboardService
}

func NewDashboardController(dashboardService service.DashboardService) *dashboardController {
	return &dashboardController{dashboardService: dashboardService}
}

func (c *dashboardController) GetDashboardData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var startDate, endDate *time.Time

	if sd := r.URL.Query().Get("start_date"); sd != "" {
		t, err := time.Parse("2006-01-02", sd)
		if err != nil {
			httperror.BadRequest("invalid start_date format, use YYYY-MM-DD", err).JSONRespond(w)
			return
		}
		startDate = &t
	}

	if ed := r.URL.Query().Get("end_date"); ed != "" {
		t, err := time.Parse("2006-01-02", ed)
		if err != nil {
			httperror.BadRequest("invalid end_date format, use YYYY-MM-DD", err).JSONRespond(w)
			return
		}
		// Set to end of day
		eod := t.Add(24*time.Hour - time.Second)
		endDate = &eod
	}

	data, err := c.dashboardService.GetDashboardData(ctx, startDate, endDate)
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

func (c *dashboardController) GetMyPermissions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	perms, ok := ctx.Value(constant.PermissionsSessionKey).(map[string]bool)
	if !ok || perms == nil {
		perms = make(map[string]bool)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(perms)
}
