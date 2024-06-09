package seller

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/dikletscode/isyana-store/db"
	"github.com/dikletscode/isyana-store/pkg/httperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Product struct {
	Id          string     `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	Price       int        `json:"price"`
	Stock       int        `json:"stock"`
	CategoryId  *int       `json:"category_id"`
	SellerId    string     `json:"seller_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"-"`
}

type response struct {
	Status string             `json:"status"`
	Data   *Product           `json:"data"`
	Errors *httperrors.Errors `json:"errors"`
}

type responseArr struct {
	Status string             `json:"status"`
	Data   []Product          `json:"data"`
	Errors *httperrors.Errors `json:"errors"`
}

func postProduct(sellerId string, product Product) response {
	if len(product.Name) <= 5 || len(*product.Description) >= 200 || len(product.Name) >= 100 || product.Price < 100 || product.Price > 100000000 || product.Stock < 0 {
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

	query := `INSERT INTO products (id ,name, description, price, stock, seller_id) VALUES (@id, @name, @description, @price, @stock, @sellerId)`
	id := uuid.New()
	product.Id = id.String()
	product.SellerId = sellerId

	args := pgx.NamedArgs{
		"id":          id,
		"name":        product.Name,
		"description": product.Description,
		"price":       product.Price,
		"stock":       product.Stock,
		"sellerId":    sellerId,
	}
	_, err := db.PG.Exec(context.Background(), query, args)

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
		Data:   &product,
		Errors: nil,
	}

}

func updateProduct(jwtId string, product Product) response {
	if len(product.Id) <= 0 {
		return response{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    400,
				Message: "Bad Request: Missing id",
			},
		}
	}
	_, err := uuid.Parse(product.Id)

	if err != nil {
		return response{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    400,
				Message: "Bad Request: Product id is invalid",
			},
		}
	}

	if len(product.Name) <= 5 || len(*product.Description) >= 200 || len(product.Name) >= 100 || product.Price < 100 || product.Price > 100000000 || product.Stock < 0 {
		return response{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    400,
				Message: "Bad Request: Invalid input data",
			},
		}
	}

	query := `UPDATE products SET 
	name=@name, description=@description, price=@price, stock=@stock
	where id=@id AND seller_id=@sellerId
	`

	args := pgx.NamedArgs{
		"id":          product.Id,
		"name":        product.Name,
		"description": product.Description,
		"price":       product.Price,
		"stock":       product.Stock,
		"sellerId":    jwtId,
	}
	comTag, err := db.PG.Exec(context.Background(), query, args)

	if comTag.RowsAffected() == 0 {
		return response{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    400,
				Message: "No rows were updated.",
			},
		}
	}
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
		Data:   &product,
		Errors: nil,
	}

}

func getProductById(sellerId string) response {
	var query string
	var rows pgx.Rows
	var err error

	query = `SELECT * FROM products where id = $1`
	rows, err = db.PG.Query(context.Background(), query, sellerId)

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

	product, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[Product])

	if err != nil {
		log.Println(err.Error())

		if err == pgx.ErrNoRows {
			return response{
				Status: "sucess",
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

func getProducts(categoryId string) responseArr {
	var query string
	var err error
	var rows pgx.Rows

	if categoryId == "" {
		query = `SELECT * FROM products`
		rows, err = db.PG.Query(context.Background(), query)
	} else {
		_, err = uuid.Parse(categoryId)
		if err != nil {
			return responseArr{
				Status: "failed",
				Data:   nil,
				Errors: &httperrors.Errors{
					Code:    400,
					Message: `Invalid category id`,
				},
			}
		}
		query = `SELECT * FROM products where category_id = $1`
		rows, err = db.PG.Query(context.Background(), query, categoryId)
	}

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

	product, err := pgx.CollectRows(rows, pgx.RowToStructByPos[Product])

	if err != nil {
		log.Println(err.Error())
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			fmt.Printf(" vvvvvv %v ", pgErr)
			if pgErr.Code == "22P02" {
				return responseArr{
					Status: "failed",
					Data:   nil,
					Errors: &httperrors.Errors{
						Code:    409,
						Message: "Username already exists. Please use a different username address.",
					},
				}
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
