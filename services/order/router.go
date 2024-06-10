package order

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dikletscode/isyana-store/middleware"
	"github.com/dikletscode/isyana-store/pkg/httperrors"
)

func SellerRouter() {
	http.Handle("/order", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodPost {

			claims := middleware.UserFromContext(r.Context())

			jwtUserID, ok := claims["jti"].(string)

			decoder := json.NewDecoder(r.Body)
			var incomingOrder order
			var resp response
			err := decoder.Decode(&incomingOrder)
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
			purchaseStatus := "IN_CART"
			purchaseSource := "cart"
			incomingOrder.UserId = &jwtUserID
			incomingOrder.PurchaseStatus = purchaseStatus
			incomingOrder.PurchaseSource = purchaseSource
			resp = addToOrder(incomingOrder)

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

			var resp responseArr

			claims := middleware.UserFromContext(r.Context())

			jwtUserID, ok := claims["jti"].(string)
			if !ok {
				// Handle the case where "jti" is not a string
				resp = responseArr{
					Status: "failed",
					Data:   nil,
					Errors: &httperrors.Errors{
						Code:    400,
						Message: "Bad Request: Invalid input data",
					},
				}
			}
			resp = getMyOrders(jwtUserID)

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

	http.Handle("/order/", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			var incomingOrder order
			incomingOrder.Id = breakUrl[len(breakUrl)-1]
			incomingOrder.UserId = &jwtUserID

			err := decoder.Decode(&incomingOrder)
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
			resp = updateOrder(incomingOrder)

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
			claims := middleware.UserFromContext(r.Context())

			jwtUserID, ok := claims["jti"].(string)
			if !ok {
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
			orderId := breakUrl[len(breakUrl)-1]
			resp = getMyOrderById(jwtUserID, orderId)

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
