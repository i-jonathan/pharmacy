package main

import (
	"log"
	"pharmacy/adapter/http/router"
	"pharmacy/repository"
)

func main() {
	_, err := repository.InitStore()
	if err != nil {
		log.Fatal(err)
	}

	_ = router.InitRouter()
}
