package utils

import (
	"encoding/json"
	"net/http"
)

type CustomError struct{
	Code int `json:"code"`
	Title string `json:"title"`
	Detail string `json:"detail"`
}

func HandleError(w http.ResponseWriter, errors []CustomError, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errors)
}