package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"pharmacy/adapter/http/middleware"
	"pharmacy/adapter/http/router"
	"pharmacy/config"
	"pharmacy/repository"
	"pharmacy/service"
	"strings"
	"time"
)

//go:embed template/*.html
var templateFS embed.FS

//go:embed template/static/**
var embeddedStatic embed.FS

var tmpl *template.Template

func main() {
	_ = config.Conf
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer file.Close()

	log.SetOutput(file)

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

	saleService := service.NewSaleService(store)
	saleRouter := router.InitSalesRouter(saleService, tmpl)
	r.Handle("/sales/", middleware.AuthMiddleware(saleRouter))

	middlewareStack := middleware.CreateStack(
		// middleware.CSRFMiddleware,
		middleware.Logging,
	)

	server := http.Server{
		Addr:    "0.0.0.0:8000",
		Handler: middlewareStack(r),
	}
	log.Println("Listening on port 8000...")
	server.ListenAndServe()
}

func parseTemplates() {
	var err error
	tmpl, err = template.New("").Funcs(template.FuncMap{
		"formatDate": func(v any) string {
			switch t := v.(type) {
			case time.Time:
				if t.IsZero() {
					return "-"
				}
				return t.Format("January 2, 2006")
			case *time.Time:
				if t == nil || t.IsZero() {
					return "-"
				}
				return t.Format("January 2, 2006")
			default:
				return "-"
			}
		},
		"formatPrice": func(v int) string {
			naira := float64(v) / 100.0
			s := fmt.Sprintf("%.2f", naira) // e.g. "12500.00"

			// Insert commas manually
			parts := strings.Split(s, ".")
			intPart := parts[0]
			decPart := parts[1]

			// walk from the right, insert commas every 3 digits
			var out []byte
			for i, d := range intPart {
				if (len(intPart)-i)%3 == 0 && i != 0 {
					out = append(out, ',')
				}
				out = append(out, byte(d))
			}

			return string(out) + "." + decPart
		},
		"hasPerm": func(perms map[string]bool, permissionKey string) bool {
			return perms[permissionKey]
		},
		"toJSON": func(v any) template.JS {
			b, err := json.Marshal(v)
			if err != nil {
				log.Println(err)
				return template.JS("null")
			}
			return template.JS(b)
		},
	}).ParseFS(templateFS, "template/*.html")
	if err != nil {
		panic("failed to parse templates: " + err.Error())
	}
}
