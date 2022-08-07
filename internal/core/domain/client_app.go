package domain

import (
	"encoding/hex"
	"math/rand"
	"strconv"
	"strings"
	"time"

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

	requestTokenBytes := make([]byte, 32)
	secretBytes := make([]byte, 40)

	rand.Read(requestTokenBytes)
	rand.Read(secretBytes)

	requestToken := hex.EncodeToString(requestTokenBytes)
	secret := hex.EncodeToString(secretBytes)

	return &ClientApp{
		Id:           clientId,
		AppName:      appName,
		RequestToken: requestToken,
		Secret:       secret,
	}
}
