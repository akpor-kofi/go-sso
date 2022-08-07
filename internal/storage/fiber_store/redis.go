package fiber_store

import (
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/redis"
	"time"
)

var (
	Store *session.Store
)

func ConnectRedisStore(host string, port int) {
	storage := redis.New(redis.Config{
		Host: host,
		Port: port,
	})

	Store = session.New(session.Config{
		Storage:    storage,
		Expiration: 2 * time.Hour,
	})
}
