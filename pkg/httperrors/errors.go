package httperrors

import (
	"encoding/json"
	"net/http"
)

const C500 = "Oops! Something went wrong. We're working to fix the issue. Please try again later."

type Errors struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Response struct {
	Status string  `json:"status"`
	Data   *int    `json:"data"`
	Errors *Errors `json:"errors"`
}

func HandleError(w http.ResponseWriter, errors Response, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errors)
}
