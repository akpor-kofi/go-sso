package rest

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/template/html"
)

func FiberApp() *fiber.App {
	engine := html.New("./internal/http/views", ".html")
	cookieKey := os.Getenv("DECRYPT_COOKIE_KEY")

	app := fiber.New(fiber.Config{
		AppName:           "sso server",
		Views:             engine,
		PassLocalsToViews: true,
	})

	app.Static("/static", "./internal/http/public")

	app.Use(limiter.New(limiter.Config{
		Max: 50,
	}))

	app.Use(encryptcookie.New(encryptcookie.Config{
		Key: cookieKey,
	}))

	app.Use(compress.New())
	app.Use(cors.New())

	api := app.Group("/api")
	v1 := api.Group("/v1")

	v1.Route("/oauth", oauthRoutes)
	v1.Route("/auth", authRoutes)
	v1.Route("/users", userRoutes)
	v1.Route("/clientApp", clientAppRoutes)
	v1.Route("/companies", companyRoutes)

	return app
}
