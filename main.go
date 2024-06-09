package main

import (
	"fmt"
	"net/http"

	"github.com/dikletscode/isyana-store/db"
	"github.com/dikletscode/isyana-store/pkg/utils"
	"github.com/dikletscode/isyana-store/services/auth"
	"github.com/dikletscode/isyana-store/services/order"
	"github.com/dikletscode/isyana-store/services/seller"
	"github.com/dikletscode/isyana-store/services/transaction"
)

func main() {
	utils.LoadEnvFile()
	db.DBConnect()

	auth.AuthRouters()
	order.SellerRouter()
	seller.SellerRouter()
	seller.VocuherRoute()
	transaction.SellerRouter()

	err := http.ListenAndServe(":5000", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}

}
