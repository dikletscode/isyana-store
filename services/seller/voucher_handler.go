package seller

import (
	"context"
	"log"
	"time"

	"github.com/dikletscode/isyana-store/db"
	"github.com/dikletscode/isyana-store/pkg/httperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type voucherType struct {
	Id                 string     `json:"id"`
	Name               string     `json:"name"`
	Description        *string    `json:"description"`
	Type               string     `json:"type"` /** V0S = SINGLE  Product discount V0M = MULTIPLE Products discount V0C = COMBINE = Some Product discount **/
	Status             string     `json:"status"`
	DiscountPercentage float64    `json:"discount_percentage"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"-"`
}

type responseVoucher struct {
	Status string             `json:"status"`
	Data   *voucherType       `json:"data"`
	Errors *httperrors.Errors `json:"errors"`
}

type responseVoucherArr struct {
	Status string             `json:"status"`
	Data   *[]voucherType     `json:"data"`
	Errors *httperrors.Errors `json:"errors"`
}

func isValidType(typeVoucher string) bool {
	if typeVoucher == "VOS" || typeVoucher == "VOM" || typeVoucher == "VOC" {
		return true
	}
	return false
}

func postVoucher(voucher voucherType) responseVoucher {
	if len(voucher.Name) <= 5 || len(*voucher.Description) >= 200 || !isValidType(voucher.Type) {
		// log.Println(err.Error())
		return responseVoucher{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    400,
				Message: "Bad Request: Invalid input data",
			},
		}
	}

	query := `INSERT INTO vouchers (id ,name, description, type, status, discount_percentage) VALUES (@id, @name, @description, @type, @status, @discountPercentage)`
	id := uuid.New()
	voucher.Id = id.String()

	args := pgx.NamedArgs{
		"id":                 id,
		"name":               voucher.Name,
		"description":        voucher.Description,
		"type":               voucher.Type,
		"status":             voucher.Status,
		"discountPercentage": voucher.DiscountPercentage,
	}
	_, err := db.PG.Exec(context.Background(), query, args)

	if err != nil {
		log.Println(err.Error())

		return responseVoucher{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    500,
				Message: httperrors.C500,
			},
		}
	}
	return responseVoucher{
		Status: "success",
		Data:   &voucher,
		Errors: nil,
	}
}

func putVoucher(voucher voucherType) responseVoucher {
	if len(voucher.Name) <= 5 || len(*voucher.Description) >= 200 || !isValidType(voucher.Type) {
		// log.Println(err.Error())
		return responseVoucher{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    400,
				Message: "Bad Request: Invalid input data",
			},
		}
	}

	query := `UPDATE vouchers SET 
	name=@name, description=@description, type=@type, status=@status, 
	discount_percentage=@discountPercentage where id=@id`

	args := pgx.NamedArgs{
		"id":                 voucher.Id,
		"name":               voucher.Name,
		"description":        voucher.Description,
		"type":               voucher.Type,
		"status":             voucher.Status,
		"discountPercentage": voucher.DiscountPercentage,
	}
	_, err := db.PG.Exec(context.Background(), query, args)

	if err != nil {
		log.Println(err.Error())

		return responseVoucher{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    500,
				Message: httperrors.C500,
			},
		}
	}
	return responseVoucher{
		Status: "success",
		Data:   &voucher,
		Errors: nil,
	}
}

func getAllVoucher() responseVoucherArr {

	query := `SELECT id,name, description,status,discount_percentage,created_at,updated_at FROM vouchers`
	rows, err := db.PG.Query(context.Background(), query)
	var vouchers []voucherType
	if err != nil {
		log.Println(err.Error())

		return responseVoucherArr{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    500,
				Message: httperrors.C500,
			},
		}
	}

	for rows.Next() {
		var voucher voucherType
		err = rows.Scan(&voucher.Id, &voucher.Name, &voucher.Description, &voucher.Status, &voucher.DiscountPercentage, &voucher.CreatedAt, &voucher.UpdatedAt)

		if err != nil {
			log.Println(err.Error())

			return responseVoucherArr{
				Status: "failed",
				Data:   nil,
				Errors: &httperrors.Errors{
					Code:    500,
					Message: httperrors.C500,
				},
			}
		}
		log.Println(voucher)
		vouchers = append(vouchers, voucher)
	}

	return responseVoucherArr{
		Status: "success",
		Data:   &vouchers,
		Errors: nil,
	}
}
