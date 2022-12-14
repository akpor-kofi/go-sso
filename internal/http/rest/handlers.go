package rest

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go-sso/internal/core/domain"
	"go-sso/internal/core/ports"
	email "go-sso/internal/mailing"
	"go-sso/internal/storage/fiber_store"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/nacl/auth"
)

var cb = context.Background()

type HttpHandler struct {
	userService      ports.UserService
	companyService   ports.CompanyService
	clientAppService ports.ClientAppService
	contentStorage   ports.ContentStorage
}

func NewHttpHandler(userService ports.UserService, companyService ports.CompanyService, clientAppService ports.ClientAppService, contentStorage ports.ContentStorage) *HttpHandler {
	return &HttpHandler{userService, companyService, clientAppService, contentStorage}
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
		return (err)
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
	fh, err := ctx.FormFile("imageFile")

	if err != nil {
		return err
	}

	if err := ctx.BodyParser(newUser); err != nil {
		return err
	}

	newUser.Id = utils.UUIDv4()

	_, err = http.userService.GetByEmail(newUser.Email)

	if err == nil {
		return ctx.Status(409).JSON(fiber.Map{
			"status":  "fail",
			"field":   "email",
			"message": "this email is already in use",
		})
	}

	// open multipart file
	file, err := fh.Open()
	if err != nil {
		return err
	}

	//upload file i.e image
	imageUrl, err := contentStorage.Upload(file, newUser.Id)

	if err != nil {
		return err
	}
	newUser.Image = imageUrl

	// send a verification email
	// keep the user details in redis with an expiry of 15 minutes
	b, err := json.Marshal(newUser)
	if err != nil {
		return err
	}
	fiber_store.Store.Storage.Set("user:signingup:"+newUser.Id, b, 5*time.Minute)

	e := email.New(newUser.Email, "verify signup")
	body := fmt.Sprintf("please click on the link: %s://%s/verify-signup?id=%s to verify your account. You have 15 minutes", ctx.Protocol(), ctx.Hostname(), newUser.Id)
	err = e.Send(body, "text/plain")

	if err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"status": "success",
	})
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
		return fiber.NewError(401, "you're not logged in")
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
	rand.Seed(time.Now().UnixNano())
	rand.Read(applicationCodeBytes)

	applicationCode := hex.EncodeToString(applicationCodeBytes)
	codeExpiration := 5 * time.Minute

	fiber_store.Store.Storage.Set(applicationCode, []byte(userId), codeExpiration)

	return ctx.Redirect(redirectUri + "?code=" + applicationCode + "&expiresAt=" + strconv.FormatInt(time.Now().Add(10*time.Minute).UnixMilli(), 10))
}

func (http *HttpHandler) getUserData(ctx *fiber.Ctx) error {
	appCode := ctx.Query("code")
	// clientId := ctx.Query("clientId")

	// client, err := http.clientAppService.Get(clientId)

	// if err != nil {
	// 	return err
	// }

	userIdBytes, err := fiber_store.Store.Storage.Get(appCode)

	if err != nil {
		return err
	}

	user, err := http.userService.Get(string(userIdBytes))

	//removing jwt functionality now cause of java jwt seem not to work

	// jwt sign this shit
	// claims := jwt.MapClaims{
	// 	"ventisId": user.Id,
	// 	"name":     user.Name,
	// 	"email":    user.Email,
	// 	"image":    user.Image,
	// 	"dob":      user.Dob,
	// }

	// token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// t, err := token.SignedString([]byte(client.Secret))

	// if err != nil {
	// 	return err
	// }

	// return ctx.SendString(t)

	return ctx.Status(fiber.StatusOK).JSON(*user)
}

func (http *HttpHandler) forgotPassword(ctx *fiber.Ctx) error {
	body := &struct {
		Email string `json:"email" form:"email"`
	}{}

	err := ctx.BodyParser(body)
	if err != nil {
		return err
	}

	existingUser, err := http.userService.GetByEmail(body.Email)
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"status":  "fail",
			"message": "invalid email",
		})
	}

	//get email resetToken
	resetTokenBytes := make([]byte, 32)
	rand.Seed(time.Now().UnixNano())
	rand.Read(resetTokenBytes)
	resetToken := hex.EncodeToString(resetTokenBytes)

	// this algo not working for some dumb shit
	err = http.userService.UpdateResetToken(body.Email, resetToken)
	if err != nil {
		return err
	}

	fiber_store.Store.Storage.Set(fmt.Sprintf("password:reset:%s", resetToken), []byte(existingUser.Id), 10*time.Minute)
	// instead i'll be using redis

	// arrange the sending of email
	to := body.Email
	e := email.New(to, "forgot password")

	message := fmt.Sprintf("%s://%s/reset-password?token=%s", ctx.Protocol(), ctx.Hostname(), resetToken)

	if err = e.Send(message, "text/plain"); err != nil {
		return err
	}

	return ctx.Status(200).JSON(fiber.Map{"status": "success"})
}

func (http *HttpHandler) resetPassword(ctx *fiber.Ctx) error {
	token := ctx.Params("token")
	key := fmt.Sprintf("password:reset:%s", token)

	form := &struct {
		Password        string `json:"password" form:"password"`
		ConfirmPassword string `json:"confirmPassword" form:"confirmPassword"`
	}{}

	if err := ctx.BodyParser(form); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	if form.Password != form.ConfirmPassword {
		err := fmt.Errorf("password is not the same")
		return ctx.Status(fiber.StatusConflict).JSON(err.Error())
	}

	b, err := fiber_store.Store.Storage.Get(key)

	if err != nil || len(b) == 0 {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	hashedByte, _ := bcrypt.GenerateFromPassword([]byte(form.Password), 12)
	password := string(hashedByte)

	if err = http.userService.UpdatePassword(string(b), password); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	return ctx.SendStatus(fiber.StatusOK)
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

func (http *HttpHandler) resetPasswordForm(ctx *fiber.Ctx) error {
	token := ctx.Query("token")

	key := fmt.Sprintf("password:reset:%s", token)

	b, err := fiber_store.Store.Storage.Get(key)

	if err != nil || len(b) == 0 {
		return ctx.Render("404", fiber.Map{
			"Title": "Ventis | Page Not Found",
		})
	}

	return ctx.Status(200).Render("resetPassword", fiber.Map{
		"Title": "Ventis | Reset Password",
	})
}

func (http *HttpHandler) verifySignup(ctx *fiber.Ctx) error {
	id := ctx.Query("id")

	key := "user:signingup:" + id

	user := new(domain.User)

	b, err := fiber_store.Store.Storage.Get(key)

	if err != nil {
		return err
	}

	err = json.Unmarshal(b, user)

	if err != nil {
		return ctx.SendString("invalid or expired link")
	}

	_, err = http.userService.New(user)

	if err != nil {
		return err
	}

	createSignToken(user, ctx, 201)
	return ctx.Render("verifySignup", fiber.Map{"Title": "Sign up verified"})
}

func (http *HttpHandler) validateUserBody(ctx *fiber.Ctx) error {
	user := new(domain.User)

	if err := ctx.BodyParser(user); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": err.Error()})
	}

	errors := domain.UserValidation(*user)
	if errors != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": errors})
	}

	return ctx.Next()
}
