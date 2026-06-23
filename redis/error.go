package redis

import (
	"errors"
	"github.com/redis/go-redis/v9"
)

func IsKeyNotFound(err error) bool {
	return errors.Is(err, redis.Nil)
}

func IsTimeout(err error) bool {
	return errors.Is(err, redis.ErrClosed)
}
