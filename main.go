package main

import (
	"log"
	"net/http"
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

	log.Println("Listening on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", router))
}
