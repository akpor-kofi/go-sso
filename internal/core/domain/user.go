package domain

import (
	"encoding/hex"
	"math/rand"
	"time"

	"github.com/go-playground/validator"
	"golang.org/x/crypto/bcrypt"
)

const (
	errv = iota
	new
	update
)

type ErrorResponse struct {
	FailedField string
	Tag         string
	Value       string
}

type User struct {
	Id             string `json:"id"`
	Name           string `json:"name" form:"name" validate:"required"`
	Email          string `json:"email" form:"email" validate:"required,email"`
	Image          string `json:"image"`
	Dob            string `json:"dob" form:"dob" validate:"required"`
	Password       string `json:"password" form:"password" validate:"required,min=8"`
	EmployeeId     string `json:"employeeId"`
	PhoneNumber    string `json:"phoneNumber" phoneNumber:"phoneNumber"`
	CreatedAt      int64  `json:"createdAt"`
	UpdatedAt      int64  `json:"updatedAt"`
	ResetToken     string `json:"resetToken"`
	ResetExpiresAt int64  `json:"resetExpiresAt"`
	Version        int    `json:"version"`
	Active         bool   `json:"active"`
	// TODO: reset token property
}

// Pre hooks for the user
func (u *User) Pre(hook, writeType string) {
	var w int
	switch writeType {
	case "new":
		w = new
	case "update":
		w = update
	}

	switch hook {
	case "save":
		// bcrypt password
		hashedByte, _ := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
		u.Password = string(hashedByte)

		// createdAt and updatedAt
		if w == new {
			u.CreatedAt = time.Now().UnixMilli()
			u.UpdatedAt = time.Now().UnixMilli()
			u.Active = true
		}

		if w == update {
			u.UpdatedAt = time.Now().UnixMilli()
			u.Version++
		}

		// setting reset token
		resetTokenBytes := make([]byte, 32)
		rand.Seed(time.Now().UnixNano())
		rand.Read(resetTokenBytes)

		u.ResetToken = hex.EncodeToString(resetTokenBytes)
	}
}

func (u *User) Post(hook string) {
	switch hook {
	case "find":
		u.Password = ""
	}
}

func (u *User) CompareHashPassword(hashedPassword, password string) error {
	hashedPasswordBytes := []byte(hashedPassword)
	inputPasswordBytes := []byte(password)
	if err := bcrypt.CompareHashAndPassword(hashedPasswordBytes, inputPasswordBytes); err != nil {
		return err
	}

	return nil
}

func UserValidation(user User) []*ErrorResponse {
	var validate = validator.New()
	var errors []*ErrorResponse
	err := validate.Struct(user)
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
