package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9/maintnotifications"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig redis配置信息
type RedisConfig struct {
	Host     string `mapstructure:"host" json:"host" yaml:"host"`                // 地址
	Port     int    `mapstructure:"port" json:"port" yaml:"port"`                // 端口
	Password string `mapstructure:"password" json:"password" yaml:"password"`    // 密码
	DB       int    `mapstructure:"db" json:"db" yaml:"db"`                      // 可选 默认0
	PoolSize int    `mapstructure:"pool_size" json:"pool_size" yaml:"pool_size"` // 可选 默认1
}

func (r RedisConfig) String() string {
	return fmt.Sprintf("redis://:%s@%s:%d?db=%d&pool_size=%d", r.Password, r.Host, r.Port, r.DB, r.PoolSize)
}

type RedisClient = redis.Cmdable

// type RedisClient = *redis.Client

// InitRedis
// redis有单机、集群、哨兵模式，这些模式的实例Client类型均不一致
// 但都实现了业务操作接口 redis.Cmdable
// 这里使用 redis.Cmdable 方便后续切换redis模式不用改动业务代码
func InitRedis(c RedisConfig) (RedisClient, error) {
	if c.PoolSize <= 0 {
		c.PoolSize = 1
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:        fmt.Sprintf("%s:%d", c.Host, c.Port),
		Password:    c.Password,
		DB:          c.DB,
		PoolTimeout: 60 * time.Second,

		PoolSize: c.PoolSize,
		MaintNotificationsConfig: &maintnotifications.Config{
			Mode: maintnotifications.ModeDisabled,
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		// log.Printf("connect redis fail!!! err: %v config: %s", err, c.String())
		return nil, err
	}

	// log.Printf("connect redis success! config: %s", c.String())
	return redisClient, nil
}
