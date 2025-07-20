package main

import (
	"log"
	"pharmacy/adapter/http/router"
	"pharmacy/config"
	"pharmacy/repository"
)

func main() {
	err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	
	_, err = repository.InitStore()
	if err != nil {
		log.Fatal(err)
	}
	
	_ = router.InitRouter()
}