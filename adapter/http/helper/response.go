package helper

import (
	"encoding/json"
	"log"
	"net/http"
	"pharmacy/httperror"
)

func JSONResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println(err)
		httperror.ServerError("failed to encode response data", err)
	}
}
