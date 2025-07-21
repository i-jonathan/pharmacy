package main

import (
	"fmt"
	"log"
	"pharmacy/adapter/http/router"
	"pharmacy/config"
	"pharmacy/repository"
	"pharmacy/service"
)

func main() {
	store, err := repository.InitStore()
	fmt.Println(config.Conf)
	if err != nil {
		log.Fatal(err)
	}

	userService := service.NewUserService(store)
	userController := router.InitUserRouter(userService)

	router := router.InitRouter()
	router.Handle("/user/", userController)
}
