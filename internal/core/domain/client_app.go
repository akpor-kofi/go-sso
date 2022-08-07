package domain

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	_ "github.com/golang-jwt/jwt/v4"
)

type ClientApp struct {
	Id           string `json:"Id"`
	AppName      string `json:"appName"`
	RequestToken string `json:"requestToken"`
	Secret       string `json:"secret"`

	// TODO: finish this
}

func New(appName string) *ClientApp {
	appNameSlug := strings.ReplaceAll(appName, " ", "-")
	randomNum := rand.Intn(10)
	clientId := strconv.FormatInt(time.Now().UnixMilli(), 10) + "-" + strconv.FormatInt(int64(randomNum), 10) + "." + appNameSlug + ".ventis-inc.client-user-authorizer.com"

	secretBytes := make([]byte, 40)

	claims := jwt.MapClaims{
		"id":   clientId,
		"type": "request",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString(secretBytes)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(t, "jwt  token here")

	rand.Read(secretBytes)

	secret := hex.EncodeToString(secretBytes)

	return &ClientApp{
		Id:           clientId,
		AppName:      appName,
		RequestToken: t,
		Secret:       secret,
	}
}
