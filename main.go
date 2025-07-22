package main

import (
	"log"
	"net/http"
	"pharmacy/adapter/http/middleware"
	"pharmacy/adapter/http/router"
	"pharmacy/repository"
	"pharmacy/service"
)

func main() {
	store, err := repository.InitStore()
	if err != nil {
		log.Fatal(err)
	}

	userService := service.NewUserService(store)
	userController := router.InitUserRouter(userService)

	router := router.InitRouter()
	router.Handle("/user/", userController)

	server := http.Server{
		Addr: ":8000",
		Handler: middleware.Logging(router),
	}
	log.Println("Listening on port 8000...")
	server.ListenAndServe()
}
