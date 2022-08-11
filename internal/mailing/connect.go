package email

import "gopkg.in/gomail.v2"

var (
	D *gomail.Dialer
)

func ConnectToEmailService(host string, port int, from, password string) {
	D = gomail.NewDialer(host, port, from, password)
}
