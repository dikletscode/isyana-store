package auth

import "time"

type userLogin struct {
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
}

type user struct {
	Id              string    `json:"id"`
	FullName        *string   `json:"full_name,omitempty"`
	Username        string    `json:"username"`
	Photo           *string   `json:"photo,omitempty"`
	ShippingAddress *string   `json:"shipping_address,omitempty"`
	UserType        *string   `json:"user_type,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
