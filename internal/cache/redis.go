package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/CodingFervor/workflow-engine/internal/config"
	"github.com/CodingFervor/workflow-engine/pkg/logger"
)

var RDB *redis.Client

func Connect(cfg config.RedisConfig) error {
	RDB = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     10,
		MinIdleConns: 5,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := RDB.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}
	logger.Info("redis connected", "addr", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
	return nil
}

func Close() {
	if RDB != nil {
		RDB.Close()
		logger.Info("redis connection closed")
	}
}

func Set(ctx context.Context, key string, val interface{}, ttl time.Duration) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return RDB.Set(ctx, key, data, ttl).Err()
}

func Get(ctx context.Context, key string, dest interface{}) error {
	data, err := RDB.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func Delete(ctx context.Context, keys ...string) error {
	return RDB.Del(ctx, keys...).Err()
}
