package rest

import (
	"context"
	"encoding/hex"
	"fmt"
	"go-sso/internal/core/domain"
	"go-sso/internal/core/ports"
	"go-sso/internal/email"
	"go-sso/internal/storage/fiber_store"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/nacl/auth"
)

var cb = context.Background()

type HttpHandler struct {
	userService      ports.UserService
	companyService   ports.CompanyService
	clientAppService ports.ClientAppService
}

func NewHttpHandler(userService ports.UserService, companyService ports.CompanyService, clientAppService ports.ClientAppService) *HttpHandler {
	return &HttpHandler{userService, companyService, clientAppService}
}

func createSignToken(user *domain.User, ctx *fiber.Ctx, statusCode int) error {
	user.Post("find")
	sess := newSession(user.Id)

	err := sess.saveSession()
	if err != nil {
		return err
	}
	cookie := new(fiber.Cookie)
	cookie.Name = "auth"
	cookie.Value = sess.sessionToCookie()
	cookie.Expires = time.Now().Add(4 * time.Hour)

	ctx.Cookie(cookie)

	// later may send token
	return ctx.Status(statusCode).JSON(user)
}

func nothing(ctx *fiber.Ctx) error {

	return ctx.SendString("nothing at the moment")
}

func (http *HttpHandler) addUser(ctx *fiber.Ctx) error {
	newUser := new(domain.User)

	if err := ctx.BodyParser(newUser); err != nil {
		return err
	}

	user, err := http.userService.New(newUser)

	if err != nil {
		panic(err)
	}

	return ctx.JSON(user)
}

func (http *HttpHandler) getAllUsers(ctx *fiber.Ctx) error {

	users, err := http.userService.GetAll()

	if err != nil {
		log.Fatal(err)
	}

	ctx.SendStatus(200)
	return ctx.JSON(users)
}

func (http *HttpHandler) getUserById(ctx *fiber.Ctx) error {
	userId := ctx.Params("id")

	user, err := http.userService.Get(userId)
	if err != nil {
		return err
	}

	ctx.SendStatus(200)

	return ctx.JSON(user)
}

func (http *HttpHandler) deleteUser(ctx *fiber.Ctx) error {
	userId := ctx.Params("id")

	err := http.userService.Delete(userId)
	if err != nil {
		return err
	}

	return ctx.SendStatus(204)
}

func (http *HttpHandler) signup(ctx *fiber.Ctx) error {
	newUser := new(domain.User)

	if err := ctx.BodyParser(newUser); err != nil {
		return err
	}

	newUser.Id = utils.UUIDv4()

	user, err := http.userService.New(newUser)

	// implement a email service

	if err != nil {
		panic(err)
	}

	return createSignToken(user, ctx, 201)
}

func (http *HttpHandler) login(ctx *fiber.Ctx) error {
	user := &struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	if err := ctx.BodyParser(user); err != nil {
		return err
	}

	if user.Email == "" || user.Password == "" {
		// return fiber.NewError(400, "please provide email and password")
		return ctx.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "please provide email and password",
		})
	}

	existingUser, err := http.userService.GetByEmail(user.Email)
	if err != nil || existingUser.CompareHashPassword(existingUser.Password, user.Password) != nil {
		// return fiber.NewError(400, "incorrect username or password")
		return ctx.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "incorrect username or password",
		})
	}

	return createSignToken(existingUser, ctx, 200)
}

func (http *HttpHandler) protect(ctx *fiber.Ctx) error {
	authToken := ctx.Cookies("auth")
	if authToken == "" {
		return fiber.NewError(401, "you're not logged in")
	}

	token := strings.Split(authToken, ":")
	sessionId, _ := hex.DecodeString(token[0])
	sig, _ := hex.DecodeString(token[1])
	privateKey := getPrivateKeyBytes()

	if !auth.Verify(sig, sessionId, &privateKey) {
		return fiber.NewError(401, "you're not logged in")
	}

	userBytes, _ := fiber_store.Store.Storage.Get(token[0])

	userId := string(userBytes)

	if userId == "" {
		fmt.Println("user session not declared")
	}

	user, err := http.userService.Get(userId)
	if err != nil {
		return err
	}
	user.Post("find")

	// like setting req.user
	c := ctx.UserContext()
	c = context.WithValue(c, "currentUser", user)
	ctx.SetUserContext(c)

	return ctx.Next()
}

func (http *HttpHandler) isLoggedIn(ctx *fiber.Ctx) error {
	authToken := ctx.Cookies("auth")
	if authToken == "" {
		return ctx.Next()
	}

	token := strings.Split(authToken, ":")
	sessionId, _ := hex.DecodeString(token[0])
	sig, _ := hex.DecodeString(token[1])
	privateKey := getPrivateKeyBytes()

	if !auth.Verify(sig, sessionId, &privateKey) {
		return ctx.Next()
	}

	userBytes, _ := fiber_store.Store.Storage.Get(token[0])

	userId := string(userBytes)

	if userId == "" {
		fmt.Println("user session not declared")
	}

	user, err := http.userService.Get(userId)
	if err != nil {
		return err
	}
	user.Post("find")

	// like setting req.user
	c := ctx.UserContext()
	c = context.WithValue(c, "currentUser", user)
	ctx.SetUserContext(c)

	return ctx.Next()
}

func (http *HttpHandler) currentUser(ctx *fiber.Ctx) error {
	c := ctx.UserContext()
	user := c.Value("currentUser").(*domain.User)

	return ctx.JSON(user)
}

func (http *HttpHandler) logout(ctx *fiber.Ctx) error {
	// 1) remove from context
	c := ctx.UserContext()
	c = context.WithValue(c, "currentUser", "")

	// 2) delete session from redis
	authToken := ctx.Cookies("auth")
	if authToken == "" {
		return fiber.NewError(401, "you're not logged in")
	}

	sessionId := strings.Split(authToken, ":")[0]
	err := fiber_store.Store.Storage.Delete(sessionId)
	if err != nil {
		return err
	}

	// 3) clear cookie
	cookie := new(fiber.Cookie)
	cookie.Name = "auth"
	cookie.Value = ""

	ctx.Cookie(cookie)

	return ctx.SendString("logged out successfully")
}

func (http *HttpHandler) addCompany(ctx *fiber.Ctx) error {
	newCompany := new(domain.Company)

	if err := ctx.BodyParser(newCompany); err != nil {
		return err
	}
	// get logged in user
	c := ctx.UserContext()
	owner := c.Value("currentUser").(*domain.User)

	company, err := http.companyService.New(newCompany, owner)
	if err != nil {
		return err
	}

	return ctx.JSON(company)
}

func (http *HttpHandler) getAllCompanies(ctx *fiber.Ctx) error {
	companies, err := http.companyService.GetAll()
	if err != nil {
		return err
	}

	ctx.SendStatus(200)
	return ctx.JSON(companies)
}

func (http *HttpHandler) getCompanyById(ctx *fiber.Ctx) error {
	companyId := ctx.Params("id")

	company, err := http.companyService.Get(companyId)
	if err != nil {
		return err
	}

	return ctx.JSON(company)
}

func (http *HttpHandler) deleteCompany(ctx *fiber.Ctx) error {
	companyId := ctx.Params("id")

	err := http.companyService.Delete(companyId)
	if err != nil {
		return err
	}

	return ctx.SendStatus(204)
}

func (http *HttpHandler) addEmployee(ctx *fiber.Ctx) error {
	body := &struct {
		Role string `json:"role"`
	}{}

	companyId := ctx.Params("id")
	userId := ctx.Params("userId")

	if err := ctx.BodyParser(body); err != nil {
		return err
	}

	c := ctx.UserContext()
	currentUser := c.Value("currentUser").(*domain.User)

	role := http.companyService.GetCompanyRole(companyId, currentUser.Id)

	if role != "owner" {
		return fiber.NewError(400, "Not Authorised")
	}

	if err := http.companyService.AddEmployee(companyId, userId, body.Role); err != nil {
		return err
	}

	return ctx.SendStatus(200)
}

func (http *HttpHandler) getUserCompanies(ctx fiber.Ctx) error {

	return ctx.SendString("not implemented yet")
}

func (http *HttpHandler) registerApplication(ctx *fiber.Ctx) error {
	form := &struct {
		AppName string `json:"appName"`
	}{}

	if err := ctx.BodyParser(form); err != nil {
		return err
	}

	c := ctx.UserContext()
	owner := c.Value("currentUser").(*domain.User)

	clientApp, err := http.clientAppService.New(form.AppName, owner)
	if err != nil {
		return err
	}

	return ctx.JSON(clientApp)
}

func (http *HttpHandler) getApp(ctx *fiber.Ctx) error {

	return ctx.SendString("nothing has been implemented")
}

func (http *HttpHandler) signinView(ctx *fiber.Ctx) error {
	c := ctx.UserContext()
	clientApp := c.Value("clientApp").(*domain.ClientApp)
	redirectUri := ctx.Query("redirectUri")

	value := c.Value("currentUser")
	if value != nil {
		currentUser := value.(*domain.User)

		return ctx.Render("authorize", fiber.Map{
			"Title":       "Ventis | authorize",
			"RedirectUri": redirectUri,
			"AppName":     clientApp.AppName,
			"Username":    currentUser.Name,
		})
	}

	fmt.Println("got here by ventis")

	return ctx.Render("signin", fiber.Map{
		"Title":       "Ventis | sign in",
		"RedirectUri": redirectUri,
		"AppName":     clientApp.AppName,
	})
}

func (http *HttpHandler) authorize(ctx *fiber.Ctx) error {
	// requestToken := ctx.GetReqHeaders()["Authorization"]

	requestToken := ctx.Query("requestToken")
	clientId := ctx.Query("clientId")

	clientApp, err := http.clientAppService.AuthorizeClientCredentials(requestToken, clientId)
	if err != nil {
		fmt.Println(err)
		return ctx.Render("404", fiber.Map{
			"Title": "Ventis | Page Not Found",
		})
	}

	// pass the clientApp details through context
	c := ctx.UserContext()
	clientContext := context.WithValue(c, "clientApp", clientApp)
	ctx.SetUserContext(clientContext)

	return ctx.Next()
}

func (http *HttpHandler) generateCodeForClient(ctx *fiber.Ctx) error {
	userId := ctx.Query("userId")
	redirectUri := ctx.Query("redirectUri")

	// clientApp := c.Value("clientApp").(*domain.ClientApp)

	applicationCodeBytes := make([]byte, 32)
	rand.Read(applicationCodeBytes)

	applicationCode := hex.EncodeToString(applicationCodeBytes)
	codeExpiration := 5 * time.Minute

	fiber_store.Store.Storage.Set(applicationCode, []byte(userId), codeExpiration)

	return ctx.JSON(fiber.Map{
		"redirectLink": redirectUri + "?code=" + applicationCode + "&expiresAt=" + strconv.FormatInt(time.Now().Add(10*time.Minute).UnixMilli(), 10),
	})
}

func (http *HttpHandler) getUserData(ctx *fiber.Ctx) error {
	appCode := ctx.Query("code")
	clientId := ctx.Query("clientId")

	client, err := http.clientAppService.Get(clientId)

	if err != nil {
		return err
	}

	userIdBytes, err := fiber_store.Store.Storage.Get(appCode)

	if err != nil {
		return err
	}

	user, err := http.userService.Get(string(userIdBytes))

	// jwt sign this shit
	claims := jwt.MapClaims{
		"ventisId": user.Id,
		"name":     user.Name,
		"email":    user.Email,
		"image":    user.Image,
		"dob":      user.Dob,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(client.Secret))

	if err != nil {
		return err
	}

	return ctx.SendString(t)
}

func (http *HttpHandler) forgotPassword(ctx *fiber.Ctx) error {
	body := &struct {
		Email string `json:"email"`
	}{}

	err := ctx.BodyParser(body)
	if err != nil {
		return err
	}

	//get email resetToken
	resetTokenBytes := make([]byte, 32)
	rand.Read(resetTokenBytes)
	resetToken := hex.EncodeToString(resetTokenBytes)

	err = http.userService.UpdateResetToken(body.Email, resetToken)
	if err != nil {
		return err
	}

	fmt.Println("reached here")

	// arrange the sending of email
	from := os.Getenv("VENTIS_EMAIL")
	to := body.Email
	e := email.NewEmail(from, to, "forgot password")

	message := fmt.Sprintf("%s://%s/api/v1/users/resetPassword/%s", ctx.Protocol(), ctx.Hostname(), resetToken)

	if err = e.Send(message, "text/plain"); err != nil {
		return err
	}

	return ctx.SendStatus(200)
}

func (http *HttpHandler) resetPassword(ctx *fiber.Ctx) error {

	return ctx.SendString("")
}

func (http *HttpHandler) signupForm(ctx *fiber.Ctx) error {
	return ctx.Status(200).Render("signup", fiber.Map{
		"Title": "Ventis | Signup",
	})
}

func (http *HttpHandler) loginForm(ctx *fiber.Ctx) error {
	return ctx.Status(200).Render("login", fiber.Map{
		"Title": "Ventis | Signin",
	})
}

func (http *HttpHandler) forgotPasswordForm(ctx *fiber.Ctx) error {
	return ctx.Status(200).Render("forgotPassword", fiber.Map{
		"Title": "Ventis | Forgot Password",
	})
}
