package fiber_store

import (
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/redis"
)

var (
	Store *session.Store
)

func ConnectRedisStore(uri string) {
	storage := redis.New(redis.Config{
		URL: uri,
	})

	Store = session.New(session.Config{
		Storage:    storage,
		Expiration: 2 * time.Hour,
	})
}
