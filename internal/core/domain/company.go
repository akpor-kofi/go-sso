package domain

import (
	"time"
)

type Company struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	CacNum      string `json:"cacNum"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
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
