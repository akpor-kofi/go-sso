package fiber_store

import (
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/redis"
)

var (
	Store *session.Store
)

func ConnectRedisStore(host, password string, port int) {
	storage := redis.New(redis.Config{
		Host:     host,
		Port:     port,
		Username: "default",
		Password: password,
	})

	Store = session.New(session.Config{
		Storage:    storage,
		Expiration: 2 * time.Hour,
	})
}
