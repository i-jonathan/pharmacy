package router

import (
	"fmt"
	"net/http"
	"pharmacy/adapter/http/controller"
	"pharmacy/service"
)

func InitDashboardRouter(dashboardService service.DashboardService) http.Handler {
	dashboardController := controller.NewDashboardController(dashboardService)
	dashboardMux := http.NewServeMux()

	// Test endpoint
	dashboardMux.HandleFunc(http.MethodGet+" /test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Dashboard router is working!")
	})

	dashboardMux.HandleFunc(http.MethodGet+" /dashboard", dashboardController.GetDashboardData)

	return http.StripPrefix("/api", dashboardMux)
}
