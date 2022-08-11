package rest

import (
	"go-sso/internal/core/services/client_app_srv"
	"go-sso/internal/core/services/company_srv"
	"go-sso/internal/core/services/user_srv"
	"go-sso/internal/storage/neo4j"
	"go-sso/internal/storage/s3_bucket"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
)

var (
	userRepository      = neo4j.NewUserStorage()
	companyRepository   = neo4j.NewCompanyStorage()
	clientAppRepository = neo4j.NewClientAppStorage()

	userService      = user_srv.New(userRepository)
	companyService   = company_srv.New(companyRepository)
	clientAppService = client_app_srv.New(clientAppRepository)

	contentStorage = s3_bucket.NewBucket()

	httpHandler = NewHttpHandler(userService, companyService, clientAppService, contentStorage)
)

func clientAppRoutes(router fiber.Router) {
	router.Use(httpHandler.protect)
	router.Post("/register", httpHandler.registerApplication)
	router.Get("/", nothing)
	router.Get("/:id", nothing)
	router.Delete("/:id", nothing)
}

func userRoutes(router fiber.Router) {
	router.Post("/forgot-password", httpHandler.forgotPassword)
	router.Patch("/reset-password/:token", httpHandler.resetPassword)
	router.Use(csrf.New())
	router.Use(httpHandler.protect)
	router.Get("/current-user", httpHandler.currentUser)
	// find a way to get companies that the user is working in already
	router.Get("/", httpHandler.getAllUsers)
	router.Post("/", httpHandler.validateUserBody, httpHandler.addUser)
	router.Get("/:id", httpHandler.getUserById)
	router.Patch("/:id", nothing)
	router.Delete("/:id", httpHandler.deleteUser)
}

func companyRoutes(router fiber.Router) {
	router.Use(httpHandler.protect)
	router.Post("/", httpHandler.addCompany)
	router.Get("/", httpHandler.getAllCompanies)
	router.Get("/:id", httpHandler.getCompanyById)
	router.Patch("/:id", nothing)
	router.Delete("/:id", httpHandler.deleteCompany)

	router.Post("/:id/add-employee/:userId", httpHandler.addEmployee)
	// add a add location route
}

func oauthRoutes(router fiber.Router) {
	oauth := router.Group("/authorize", httpHandler.authorize)
	oauth.Get("/signin", httpHandler.isLoggedIn, httpHandler.signinView)
	oauth.Get("/handshake", httpHandler.generateCodeForClient)
	router.Get("/userinfo", httpHandler.getUserData)
}

func authRoutes(router fiber.Router) {
	router.Post("/signup", httpHandler.validateUserBody, httpHandler.signup)
	router.Post("/login", httpHandler.login)
	router.Get("/logout", httpHandler.logout)
}

func viewRoutes(router fiber.Router) {
	router.Get("/signup", httpHandler.signupForm)
	router.Get("/login", httpHandler.loginForm)
	router.Get("/forgot-password", httpHandler.forgotPasswordForm)
	router.Get("/reset-password", httpHandler.resetPasswordForm)
	router.Get("/verify-signup", httpHandler.verifySignup)
}
