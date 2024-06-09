package auth

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/dikletscode/isyana-store/db"
	"github.com/dikletscode/isyana-store/pkg/httperrors"
	"github.com/dikletscode/isyana-store/pkg/validator"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type response struct {
	Status string             `json:"status"`
	Data   *user              `json:"data"`
	Errors *httperrors.Errors `json:"errors"`
}

type token struct {
	Access_token string `json:"access_token"`
}
type loginResponse struct {
	Status string             `json:"status"`
	Data   *token             `json:"data"`
	Errors *httperrors.Errors `json:"errors"`
}

func register(newUser userLogin) response {

	if validator.IsContainSymbol(newUser.Username) || validator.IsNotValidPassword(newUser.Password) {
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

	hash, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), 12)
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

	query := `INSERT INTO users (id ,username, password, created_at, updated_at) VALUES (@id, @userName, @userPassword, @createdAt, @updatedAt)`
	id := uuid.New()

	args := pgx.NamedArgs{
		"id":           id,
		"userName":     newUser.Username,
		"userPassword": string(hash),
		"createdAt":    time.Now(),
		"updatedAt":    time.Now(),
	}

	_, err = db.PG.Exec(context.Background(), query, args)

	if err != nil {
		log.Println(err.Error())
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.ConstraintName == "users_username_key" && pgErr.Code == "23505" {
				return response{
					Status: "failed",
					Data:   nil,
					Errors: &httperrors.Errors{
						Code:    409,
						Message: "Username already exists. Please use a different username address",
					},
				}
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
		Data: &user{
			Id:       id.String(),
			Username: newUser.Username,
		},
		Errors: nil,
	}

}

func login(userRequest userLogin) loginResponse {

	if validator.IsContainSymbol(userRequest.Username) || validator.IsNotValidPassword(userRequest.Password) {
		// log.Println(err.Error())
		return loginResponse{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    400,
				Message: "Bad Request: Invalid input data",
			},
		}
	}

	query := "SELECT id, username, password FROM users WHERE username = $1"

	var id string
	var username string
	var password string
	err := db.PG.QueryRow(context.Background(), query, userRequest.Username).Scan(&id, &username, &password)

	if err != nil {
		log.Println(err.Error())

		if err == pgx.ErrNoRows {
			return loginResponse{
				Status: "failed",
				Data:   nil,
				Errors: &httperrors.Errors{
					Code:    401,
					Message: "Unauthorize",
				},
			}
		}
		return loginResponse{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    500,
				Message: httperrors.C500,
			},
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(userRequest.Password))
	if err != nil {

		log.Println(err.Error())
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return loginResponse{
				Status: "failed",
				Data:   nil,
				Errors: &httperrors.Errors{
					Code:    401,
					Message: "Unauthorize",
				},
			}
		}
		return loginResponse{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    500,
				Message: httperrors.C500,
			},
		}
	}
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "test",
		ID:        id,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	mySigningKey := []byte(os.Getenv("SECRET_TOKEN"))

	signed, err := tok.SignedString(mySigningKey)

	if err != nil {
		log.Println(err.Error())
		return loginResponse{
			Status: "failed",
			Data:   nil,
			Errors: &httperrors.Errors{
				Code:    500,
				Message: httperrors.C500,
			},
		}
	}

	return loginResponse{
		Status: "success",
		Data:   &token{Access_token: signed},
		Errors: nil,
	}

}
func getProfile(claims jwt.MapClaims) response {

	query := `SELECT 
	id, full_name, username, photo, shipping_address, user_type, created_at, updated_at  
	FROM users where id = $1`

	// args := pgx.NamedArgs{
	// 	"userName": claims["jti"],
	// }
	rows, err := db.PG.Query(context.Background(), query, claims["jti"])

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
	account, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[user])

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
		Data:   &account,
		Errors: nil,
	}

}
