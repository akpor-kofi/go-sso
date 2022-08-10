package email

import (
	"gopkg.in/gomail.v2"
)

type Email struct {
	From    string
	To      string
	Subject string
}

func NewEmail(from, to, subject string) *Email {
	return &Email{
		from, to, subject,
	}
}

// text/plain | text/html
func (e *Email) Send(body string, contentType string) error {
	m := gomail.NewMessage()

	m.SetHeader("From", e.From)
	m.SetHeader("To", e.To)
	m.SetHeader("Subject", e.Subject)
	m.SetBody(contentType, body)

	if err := D.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
