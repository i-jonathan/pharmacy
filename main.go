package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"pharmacy/adapter/http/middleware"
	"pharmacy/adapter/http/router"
	"pharmacy/repository"
	"pharmacy/service"
)

//go:embed template/*.html
var templateFS embed.FS

//go:embed template/static/**
var embeddedStatic embed.FS

var tmpl *template.Template

func main() {
	parseTemplates()
	subFS, err := fs.Sub(embeddedStatic, "template/static")
	if err != nil {
		log.Fatal(err)
	}
	store, err := repository.InitStore()
	if err != nil {
		log.Fatal(err)
	}

	r := router.InitRouter()
	staticHandler := http.FileServer(http.FS(subFS))
	r.Handle("/static/", http.StripPrefix("/static/", staticHandler))

	userService := service.NewUserService(store)
	userRouter := router.InitUserRouter(userService, tmpl)
	r.Handle("/user/", userRouter)

	appRouter := router.InitAppRouter(tmpl)
	r.Handle("/app/", middleware.AuthMiddleware(appRouter))

	inventoryService := service.NewInventoryService(store)
	inventoryRouter := router.InitInventoryRouter(inventoryService, tmpl)
	r.Handle("/inventory/", middleware.AuthMiddleware(inventoryRouter))

	middlewareStack := middleware.CreateStack(
		// middleware.CSRFMiddleware,
		middleware.Logging,
	)

	server := http.Server{
		Addr:    ":8000",
		Handler: middlewareStack(r),
	}
	log.Println("Listening on port 8000...")
	server.ListenAndServe()
}

func parseTemplates() {
	var err error
	tmpl, err = template.ParseFS(templateFS, "template/*.html")
	if err != nil {
		panic("failed to parse templates: " + err.Error())
	}
}
