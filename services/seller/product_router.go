package seller

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dikletscode/isyana-store/middleware"
	"github.com/dikletscode/isyana-store/pkg/httperrors"
)

func SellerRouter() {
	http.Handle("/product", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		if r.Method == http.MethodPost {

			claims := middleware.UserFromContext(r.Context())

			jwtUserID, ok := claims["jti"].(string)

			decoder := json.NewDecoder(r.Body)
			var incomingProduct product
			var resp response
			err := decoder.Decode(&incomingProduct)
			if !ok || err != nil {
				// Handle the case where "jti" is not a string
				resp = response{
					Status: "failed",
					Data:   nil,
					Errors: &httperrors.Errors{
						Code:    400,
						Message: "Bad Request: Invalid input data",
					},
				}
			}
			resp = postProduct(jwtUserID, incomingProduct)

			if resp.Status == "success" {
				w.WriteHeader(http.StatusCreated)
			} else {
				w.WriteHeader(resp.Errors.Code)
			}
			err = json.NewEncoder(w).Encode(resp)
			if err != nil {
				http.Error(w, "Oops! Something went wrong. We're working to fix the issue. Please try again later.", 500)
			}

		} else if r.Method == http.MethodGet {

			categoryId := r.URL.Query().Get("category_id")

			resp := getProducts(categoryId)

			if resp.Status == "success" {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(resp.Errors.Code)
			}
			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				http.Error(w, "Oops! Something went wrong. We're working to fix the issue. Please try again later.", 500)
			}

		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

	}), []string{"GET"}))

	http.Handle("/product/", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			var resp response
			breakUrl := strings.Split(r.URL.Path, "/")
			if len(breakUrl) <= 2 {
				resp = response{
					Status: "failed",
					Data:   nil,
					Errors: &httperrors.Errors{
						Code:    400,
						Message: "Bad Request: Invalid URL format",
					},
				}
			}

			claims := middleware.UserFromContext(r.Context())

			jwtUserID, ok := claims["jti"].(string)

			decoder := json.NewDecoder(r.Body)
			var incomingProduct product
			incomingProduct.Id = breakUrl[len(breakUrl)-1]

			err := decoder.Decode(&incomingProduct)
			if !ok || err != nil {
				// Handle the case where "jti" is not a string
				resp = response{
					Status: "failed",
					Data:   nil,
					Errors: &httperrors.Errors{
						Code:    400,
						Message: "Bad Request: Invalid input data",
					},
				}
			}
			resp = updateProduct(jwtUserID, incomingProduct)

			if resp.Status == "success" {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(resp.Errors.Code)
			}
			err = json.NewEncoder(w).Encode(resp)
			if err != nil {
				http.Error(w, "Oops! Something went wrong. We're working to fix the issue. Please try again later.", 500)
			}

		} else if r.Method == http.MethodGet {

			var resp response
			breakUrl := strings.Split(r.URL.Path, "/")
			if len(breakUrl) <= 2 {
				resp = response{
					Status: "failed",
					Data:   nil,
					Errors: &httperrors.Errors{
						Code:    400,
						Message: "Bad Request: Invalid URL format",
					},
				}
			}

			id := breakUrl[len(breakUrl)-1]

			resp = getProductById(id)

			if resp.Status == "success" {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(resp.Errors.Code)
			}
			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				http.Error(w, "Oops! Something went wrong. We're working to fix the issue. Please try again later.", 500)
			}
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

	}), nil))

}
