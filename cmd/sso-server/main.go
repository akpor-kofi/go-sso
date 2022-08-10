package main

import (
	"fmt"
	"go-sso/internal/email"
	"go-sso/internal/http/rest"
	"go-sso/internal/storage/fiber_store"
	"go-sso/internal/storage/neo4j"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("./cmd/sso-server/.env")
	if err != nil {
		log.Fatal(err)
	}

	neo4j.ConnectToDb(os.Getenv("NEO4J_URI"), os.Getenv("NEO4J_USERNAME"), os.Getenv("NEO4J_PASSWORD"))

	port, _ := strconv.ParseInt(os.Getenv("REDIS_PORT"), 10, 16)

	fiber_store.ConnectRedisStore(os.Getenv("REDIS_HOST"), int(port))

	defer neo4j.Driver.Close()
	defer neo4j.Session.Close()

	from := os.Getenv("VENTIS_EMAIL")
	password := os.Getenv("VENTIS_PASSWORD")
	smtpHost := "smtp.gmail.com"
	smtpPort := 465
	email.ConnectToEmailService(smtpHost, smtpPort, from, password)

	app := rest.FiberApp()

	fmt.Println("starting server: ")

	err = app.Listen(":3000")
	if err != nil {
		panic(err)
	}
}
