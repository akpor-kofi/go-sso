package domain

import (
	"encoding/hex"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	errv = iota
	new
	update
)

type User struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	Image          string `json:"image"`
	Dob            string `json:"dob"`
	Password       string `json:"password"`
	EmployeeId     string `json:"employeeId"`
	PhoneNumber    string `json:"phoneNumber"`
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
