package seller

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dikletscode/isyana-store/middleware"
	"github.com/dikletscode/isyana-store/pkg/httperrors"
)

func VocuherRoute() {
	http.Handle("/voucher", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodPost {

			decoder := json.NewDecoder(r.Body)
			var voucher voucherType
			var resp responseVoucher
			err := decoder.Decode(&voucher)
			if err != nil {
				// Handle the case where "jti" is not a string
				resp = responseVoucher{
					Status: "failed",
					Data:   nil,
					Errors: &httperrors.Errors{
						Code:    400,
						Message: "Bad Request: Invalid input data",
					},
				}
			}
			resp = postVoucher(voucher)

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

			resp := getAllVoucher()

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

	http.Handle("/voucher/", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			var resp responseVoucher
			breakUrl := strings.Split(r.URL.Path, "/")
			if len(breakUrl) <= 2 {
				resp = responseVoucher{
					Status: "failed",
					Data:   nil,
					Errors: &httperrors.Errors{
						Code:    400,
						Message: "Bad Request: Invalid URL format",
					},
				}
			}

			decoder := json.NewDecoder(r.Body)
			var voucher voucherType
			voucher.Id = breakUrl[len(breakUrl)-1]

			err := decoder.Decode(&voucher)
			if err != nil {
				// Handle the case where "jti" is not a string
				resp = responseVoucher{
					Status: "failed",
					Data:   nil,
					Errors: &httperrors.Errors{
						Code:    400,
						Message: "Bad Request: Invalid input data",
					},
				}
			}
			resp = putVoucher(voucher)

			if resp.Status == "success" {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(resp.Errors.Code)
			}
			err = json.NewEncoder(w).Encode(resp)
			if err != nil {
				http.Error(w, "Oops! Something went wrong. We're working to fix the issue. Please try again later.", 500)
			}

		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

	}), nil))

}
