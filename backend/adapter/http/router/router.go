package router

import "net/http"

var router *http.ServeMux

func InitRouter() *http.ServeMux {
	router = http.NewServeMux()
	
	return router
}