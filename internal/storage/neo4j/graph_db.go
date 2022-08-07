package neo4j

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

var (
	Session neo4j.Session
	Driver  neo4j.Driver
)

func ConnectToDb(uri string, username string, password string) {
	auth := neo4j.BasicAuth(username, password, "")

	var err error
	Driver, err = neo4j.NewDriver(uri, auth)
	if err != nil {
		panic(err)
	}

	fmt.Println("connected to db")

	Session = Driver.NewSession(neo4j.SessionConfig{DatabaseName: "neo4j"})
}
