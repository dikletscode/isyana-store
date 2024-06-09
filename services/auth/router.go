package auth

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/dikletscode/isyana-store/middleware"
	"github.com/dikletscode/isyana-store/pkg/httperrors"
)

func AuthRouters() {

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		decoder := json.NewDecoder(r.Body)

		var user userLogin
		err := decoder.Decode(&user)
		resp := register(user)
		if err != nil {

			log.Println(err.Error())
			resp = response{
				Status: "failed",
				Data:   nil,
				Errors: &httperrors.Errors{
					Code:    500,
					Message: httperrors.C500,
				},
			}

		}

		if resp.Status == "success" {
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(resp.Errors.Code)
		}
		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, "Oops! Something went wrong. We're working to fix the issue. Please try again later.", 500)
		}

	})
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		decoder := json.NewDecoder(r.Body)
		var user userLogin
		err := decoder.Decode(&user)
		resp := login(user)
		if err != nil {
			log.Println(err.Error())
			resp = loginResponse{
				Status: "failed",
				Data:   nil,
				Errors: &httperrors.Errors{
					Code:    500,
					Message: httperrors.C500,
				},
			}
		}

		if resp.Status == "success" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(resp.Errors.Code)
		}
		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, "Oops! Something went wrong. We're working to fix the issue. Please try again later.", 500)
		}

	})

	http.Handle("/profile", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		claims := middleware.UserFromContext(r.Context())

		response := getProfile(claims)

		if response.Status == "success" {
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(response.Errors.Code)
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, "Oops! Something went wrong. We're working to fix the issue. Please try again later.", 500)
		}
	})))
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		type Test struct {
			name string
			Body string
		}
		response := Test{
			name: "Hello",
			Body: "hello oi",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode((response))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

}
