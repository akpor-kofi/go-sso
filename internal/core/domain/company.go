package domain

import (
	"time"

	"github.com/go-playground/validator"
)

type Company struct {
	Id          string `json:"id"`
	Name        string `json:"name" validate:"required" form:"name"`
	CacNum      string `json:"cacNum"`
	Email       string `json:"email" validate:"required,email" form:"email"`
	PhoneNumber string `json:"phoneNumber" validate:"required" form:"phoneNumber"`
	Currency    string `json:"currency"`
	Image       string `json:"image"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
	Version     int64  `json:"version"`
	Location    `json:"location"`
}

type Location struct {
	Lng         float32 `json:"lng"`
	Lat         float32 `json:"lat"`
	Address     string  `json:"address"`
	Description string  `json:"description"`
}

func (c *Company) Pre(hook string) {
	switch hook {
	case "save":
		// createdAt and updatedAt
		c.CreatedAt = time.Now().UnixMilli()
		c.UpdatedAt = time.Now().UnixMilli()

	}
}

func CompanyValidation(company Company) []*ErrorResponse {
	var validate = validator.New()
	var errors []*ErrorResponse
	err := validate.Struct(company)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element ErrorResponse
			element.FailedField = err.StructNamespace()
			element.Tag = err.Tag()
			element.Value = err.Param()
			errors = append(errors, &element)
		}
	}
	return errors
}
