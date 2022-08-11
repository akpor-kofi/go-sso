package domain

import (
	"encoding/hex"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator"
)

type ClientApp struct {
	Id           string `json:"Id"`
	AppName      string `json:"appName" validate:"required" form:"appName"`
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

	rand.Seed(time.Now().UnixNano())
	rand.Read(requestTokenBytes)
	rand.Seed(time.Now().UnixNano())
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

func ClientAppValidation(app ClientApp) []*ErrorResponse {
	var validate = validator.New()
	var errors []*ErrorResponse
	err := validate.Struct(app)
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
