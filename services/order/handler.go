package order

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/dikletscode/isyana-store/db"
	"github.com/dikletscode/isyana-store/pkg/httperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type order struct {
	Id             string    `json:"id"`
	ProductId      string    `json:"product_id"`
	UserId         *string   `json:"userId"`
	Note           *string   `json:"note"`
	PurchaseSource string    `json:"purchase_source"`
	PurchaseStatus string    `json:"purchase_status"`
	Quantity       int       `json:"quantity"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type response struct {
	Status string             `json:"status"`
	Data   *order             `json:"data"`
	Errors *httperrors.Errors `json:"errors"`
}

type responseArr struct {
	Status string             `json:"status"`
	Data   []order            `json:"data"`
	Errors *httperrors.Errors `json:"errors"`
}

func addToOrder(newOrder order) response {
	_, err := uuid.Parse(newOrder.ProductId)

	if newOrder.Quantity <= 0 || err != nil {
		// log.Println(err.Error())
		return response{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    400,
				Message: "Bad Request: Invalid input data",
			},
		}
	}
	var count int

	query := `SELECT count(*) FROM orders where user_id=$1`
	err = db.PG.QueryRow(context.Background(), query, *newOrder.UserId).Scan(&count)
	if err != nil {

		log.Println(err.Error())

		return response{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    500,
				Message: httperrors.C500,
			},
		}
	}
	if count >= 20 {
		return response{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    400,
				Message: "Cart Limit Exceeded: Please remove items to proceed.",
			},
		}
	}

	// query := `INSERT INTO orders
	// (id, product_id, user_id, note, purchase_source, purchase_status, quantity)
	// SELECT
	// @id, @productId, @userId, @note, @purchaseSource, @purchaseStatus, @quantity
	// WHERE (SELECT COUNT(*) FROM orders WHERE user_id = @userId) < 20
	// `
	var stock int
	query = `SELECT stock FROM products where  id=$1`
	err = db.PG.QueryRow(context.Background(), query, newOrder.ProductId).Scan(&stock)
	if err != nil {
		log.Println(err.Error())
		return response{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    500,
				Message: httperrors.C500,
			},
		}
	}
	if newOrder.Quantity > stock {

		return response{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    400,
				Message: "Insufficient stock for items ",
			},
		}
	}

	query = `INSERT INTO orders
	(id, product_id, user_id, note, purchase_source, purchase_status, quantity)
	VALUES
	(@id, @productId, @userId, @note, @purchaseSource, @purchaseStatus, @quantity)`

	newOrder.Id = uuid.New().String()
	args := pgx.NamedArgs{
		"id":             newOrder.Id,
		"productId":      newOrder.ProductId,
		"userId":         *newOrder.UserId,
		"note":           &newOrder.Note,
		"purchaseSource": newOrder.PurchaseSource,
		"purchaseStatus": newOrder.PurchaseStatus,
		"quantity":       newOrder.Quantity,
	}

	_, err = db.PG.Exec(context.Background(), query, args)

	// fmt.Printf("%v ==> ", data)
	if err != nil {

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.ConstraintName == "orders_product_id_user_id_key" {
				query = `UPDATE orders SET
			 quantity=@quantity
			 where product_id=@productId AND user_id=@userId `

				args := pgx.NamedArgs{

					"productId": newOrder.ProductId,
					"userId":    *newOrder.UserId,
					"quantity":  newOrder.Quantity,
				}

				_, err = db.PG.Exec(context.Background(), query, args)

				if err != nil {
					log.Println(err.Error())
					return response{
						Status: "failed",
						Data:   nil,
						Errors: &httperrors.Errors{
							Code:    500,
							Message: httperrors.C500,
						},
					}
				}
				return response{
					Status: "success",
					Data:   &newOrder,
					Errors: nil,
				}
			}
		}

		log.Println(err.Error())
		return response{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    500,
				Message: httperrors.C500,
			},
		}
	}

	return response{
		Status: "success",
		Data:   &newOrder,
		Errors: nil,
	}

}
func updateOrder(newOrder order) response {
	_, err := uuid.Parse(newOrder.Id)

	if newOrder.Quantity <= 0 || err != nil {
		// log.Println(err.Error())
		return response{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    400,
				Message: "Bad Request: Invalid input data",
			},
		}
	}

	query := `UPDATE orders SET
			  note=@note, purchase_source=@purchaseSource, purchase_status=@purchaseStatus, quantity=@quantity 
			  where product_id=@productId AND user_id=@userId `

	args := pgx.NamedArgs{
		"id":             newOrder.Id,
		"productId":      newOrder.ProductId,
		"userId":         *newOrder.UserId,
		"note":           &newOrder.Note,
		"purchaseSource": newOrder.PurchaseSource,
		"purchaseStatus": newOrder.PurchaseStatus,
		"quantity":       newOrder.Quantity,
	}

	_, err = db.PG.Exec(context.Background(), query, args)

	// fmt.Printf("%v ==> ", data)
	if err != nil {

		log.Println(err.Error())

		return response{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    500,
				Message: httperrors.C500,
			},
		}
	}

	return response{
		Status: "success",
		Data:   &newOrder,
		Errors: nil,
	}

}

func getMyOrders(userId string) responseArr {
	_, err := uuid.Parse(userId)

	if err != nil {
		// log.Println(err.Error())
		return responseArr{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    400,
				Message: "Bad Request: Invalid input data",
			},
		}
	}

	query := `SELECT * FROM orders where user_id = $1`

	rows, err := db.PG.Query(context.Background(), query, userId)
	if err != nil {
		log.Println(err.Error())

		return responseArr{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    500,
				Message: httperrors.C500,
			},
		}
	}
	product, err := pgx.CollectRows(rows, pgx.RowToStructByPos[order])

	if err != nil {
		log.Println(err.Error())
		if err == pgx.ErrNoRows {
			return responseArr{
				Status: "sucess",
				Data:   nil,
				Errors: nil,
			}
		}
		return responseArr{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    500,
				Message: httperrors.C500,
			},
		}
	}
	return responseArr{
		Status: "success",
		Data:   product,
		Errors: nil,
	}

}

func getMyOrderById(userId string, orderId string) response {
	_, err := uuid.Parse(userId)

	if err != nil {
		// log.Println(err.Error())
		return response{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    400,
				Message: "Bad Request: Invalid input data",
			},
		}
	}

	query := `SELECT * FROM orders where user_id = $1 AND id = $2`

	rows, err := db.PG.Query(context.Background(), query, userId, orderId)
	if err != nil {
		log.Println(err.Error())

		return response{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    500,
				Message: httperrors.C500,
			},
		}
	}
	product, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[order])

	if err != nil {
		log.Println(err.Error())

		if err == pgx.ErrNoRows {
			return response{
				Status: "success",
				Data:   nil,
				Errors: nil,
			}
		}

		return response{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    500,
				Message: httperrors.C500,
			},
		}
	}
	return response{
		Status: "success",
		Data:   &product,
		Errors: nil,
	}

}
