package transaction

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dikletscode/isyana-store/db"
	"github.com/dikletscode/isyana-store/pkg/httperrors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type transaction struct {
	Id               string    `json:"id"`
	Discount         float64   `json:"discount"`
	PreDiscounAmount int       `json:"pre_discount_amount"`
	FinalAmount      int       `json:"final_amount"`
	Invoice          string    `json:"invoice"`
	PaymentMethod    string    `json:"payment_method"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type detailsErr struct {
	OrderId           string `json:"order_id"`
	ProductStock      int    `json:"product_stock"`
	RequestedQuantity int    `json:"requested_quantity"`
}

type errCustom struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Details []detailsErr `json:"details"`
}

type response struct {
	Status string       `json:"status"`
	Data   *transaction `json:"data"`
	Errors *errCustom   `json:"errors"`
}

func addTransaction(newTransaction transaction, userId string, orderId []string) response {

	if len(newTransaction.PaymentMethod) <= 1 || len(orderId) <= 0 {
		return response{
			Status: "failed",
			Data:   nil,
			Errors: &errCustom{
				Code:    400,
				Message: "Bad Request: Invalid input data",
			},
		}
	}
	ctx := context.Background()
	tx, err := db.PG.Begin(ctx)

	if err != nil {
		log.Println(err.Error())
		return response{
			Status: "failed",
			Data:   nil,
			Errors: &errCustom{
				Code:    500,
				Message: httperrors.C500,
			},
		}
	}

	defer tx.Rollback(ctx)

	updateQuery := `UPDATE orders
	SET purchase_status='COMPLETED'
	WHERE purchase_status='IN_CART' AND id = ANY($1) `

	cmdTag, err := tx.Exec(context.Background(), updateQuery, orderId)
	if err != nil {
		log.Println(err.Error())
		return response{
			Status: "failed",
			Data:   nil,
			Errors: &errCustom{
				Code:    500,
				Message: httperrors.C500,
			},
		}
	}

	if cmdTag.RowsAffected() == 0 {
		tx.Rollback(ctx)
		return response{
			Status: "failed",
			Data:   nil,
			Errors: &errCustom{
				Code:    400,
				Message: "No orders found for transaction",
			},
		}
	}

	newTransaction.Id = uuid.New().String()
	args := pgx.NamedArgs{

		"discount":      0,
		"invoice":       "https://www.invoicesimple.com/wp-content/uploads/2018/06/Sample-Invoice-printable.png",
		"paymentMethod": newTransaction.PaymentMethod,
		"userId":        userId,
		"id":            newTransaction.Id,
	}

	query := `INSERT INTO transactions (id ,discount, pre_discount_amount, final_amount, invoice, payment_method)
	SELECT @id, 
	@discount, 
	total,
	total,
	@invoice,
	@paymentMethod
	FROM (
	SELECT SUM(orders.quantity*products.price) as total FROM orders JOIN products ON orders.product_id = products.id
 	WHERE orders.user_id = @userId AND orders.purchase_status='COMPLETED'  
	) as subquery
	 
	RETURNING *`
	rowsTransac := tx.QueryRow(context.Background(), query, args)

	err = rowsTransac.Scan(&newTransaction.Id, &newTransaction.Discount, &newTransaction.PreDiscounAmount, &newTransaction.FinalAmount, &newTransaction.Invoice, &newTransaction.Invoice, &newTransaction.CreatedAt, &newTransaction.UpdatedAt)
	if err != nil {
		log.Println(err.Error())
		return response{
			Status: "failed",
			Data:   nil,
			Errors: &errCustom{
				Code:    500,
				Message: httperrors.C500,
			},
		}
	}
	type item struct {
		Id            string
		OrderId       string
		TransactionId string
	}

	lenOrder := len(orderId)
	items := make([]item, 0, lenOrder)

	for _, str := range orderId {
		items = append(items, item{Id: uuid.New().String(), OrderId: str, TransactionId: newTransaction.Id})
	}

	copyCount, err := tx.CopyFrom(
		context.Background(),
		pgx.Identifier{"order_transactions"},
		[]string{"id", "orders_id", "transaction_id"},
		pgx.CopyFromSlice(len(items), func(i int) ([]any, error) {
			return []any{items[i].Id, items[i].OrderId, items[i].TransactionId}, nil
		}),
	)

	if err != nil {
		log.Println(err.Error())
		return response{
			Status: "failed",
			Data:   nil,
			Errors: &errCustom{
				Code:    500,
				Message: httperrors.C500,
			},
		}
	}

	if copyCount != int64(lenOrder) {

		log.Printf("unexpected row count: copied %d, expected %d", copyCount, lenOrder)
		tx.Rollback(ctx)
		return response{
			Status: "failed",
			Data:   nil,
			Errors: &errCustom{
				Code:    500,
				Message: httperrors.C500,
			},
		}
	}

	batch := &pgx.Batch{}
	for _, update := range orderId {
		fmt.Println(update)
		batch.Queue(`UPDATE products SET stock = stock - o.quantity FROM  orders o 
					 WHERE o.id = $1 AND products.id = o.product_id 
				     RETURNING o.id, o.quantity, stock`, update)
	}

	batchResult := tx.SendBatch(ctx, batch)

	// rows, err := tx.Query(ctx, ``, orderId)

	// defer rows.Close()

	// arr := make([]detailsErr, batch.Len())

	outOfStock := make([]detailsErr, 0, batch.Len())

	for i := 0; i < batch.Len(); i++ {
		rows, err := batchResult.Query()
		if err != nil {
			log.Println(err.Error())
			return response{
				Status: "failed",
				Data:   nil,
				Errors: &errCustom{
					Code:    500,
					Message: httperrors.C500,
				},
			}
		}
		defer rows.Close()

		for rows.Next() {
			var r detailsErr

			err := rows.Scan(&r.OrderId, &r.RequestedQuantity, &r.ProductStock)
			if err != nil {
				log.Println(err.Error())
				return response{
					Status: "failed",
					Data:   nil,
					Errors: &errCustom{
						Code:    500,
						Message: httperrors.C500,
					},
				}
			}
			fmt.Println(r.ProductStock, r.RequestedQuantity)
			if r.ProductStock < 0 {

				outOfStock = append(outOfStock, detailsErr{
					OrderId:           r.OrderId,
					ProductStock:      r.ProductStock + r.RequestedQuantity,
					RequestedQuantity: r.RequestedQuantity,
				})
			}

		}

	}
	batchResult.Close()

	if len(outOfStock) >= 1 {
		tx.Rollback(ctx)
		return response{
			Status: "failed",
			Data:   nil,
			Errors: &errCustom{
				Code:    400,
				Message: `Insufficient stock for items `,
				Details: outOfStock,
			},
		}
	}

	// if cmmTag.RowsAffected() == 0 {

	// 	log.Println("Orders quantities more than product stocks")
	// 	tx.Rollback(ctx)
	// 	return response{
	// 		Status: "failed",
	// 		Data:   nil,
	// 		Errors: &errCustom{
	// 			Code:    400,
	// 			Message: "Insufficient stock for items: ",
	// 		},
	// 	}
	// }

	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error committing transaction:", err)
		return response{
			Status: "failed",
			Data:   nil,
			Errors: &errCustom{
				Code:    500,
				Message: httperrors.C500,
			},
		}
	}

	return response{
		Status: "success",
		Data:   &newTransaction,
		Errors: nil,
	}

	// err = tx.Commit(context.Background())
	// if err != nil {
	// 	fmt.Printf("teeeeeeee 2 ")
	// 	log.Println(err.Error())
	// 	return response{
	// 		Status: "failed",
	// 		Data:   nil,
	// 		Errors: &errCustom{
	// 			Code:    500,
	// 			Message: httperrors.C500,
	// 		},
	// 	}
	// }
	// // if comTag.RowsAffected() == 0 {
	// // 	return response{
	// // 		Status: "failed",
	// // 		Data:   nil,
	// // 		Errors: &errCustom{
	// // 			Code:    400,
	// // 			Message: "Cart Limit Exceeded: Please remove items to proceed.",
	// // 		},
	// // 	}
	// // }
	// return response{
	// 	Status: "success",
	// 	Data:   &newTransaction,
	// 	Errors: nil,
	// }

}
