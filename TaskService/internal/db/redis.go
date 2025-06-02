package db

import (
	"ExpensesService/internal/model"
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

var prefix = "expense:"

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr string) *RedisCache {
	rdb := redis.NewClient(
		&redis.Options{
			Addr: addr,
		})
	return &RedisCache{client: rdb}
}

func (r *RedisCache) SetExpense(key string, value model.Expense, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, prefix+key, data, ttl).Err()
}

func (r *RedisCache) GetExpense(key string) (*model.Expense, error) {
	data, err := r.client.Get(ctx, prefix+key).Result()
	if err != nil {
		return nil, err
	}

	var expense model.Expense
	err = json.Unmarshal([]byte(data), &expense)
	if err != nil {
		return nil, err
	}
	return &expense, nil
}

func (r *RedisCache) DeleteExpense(key string) error {
	return r.client.Del(ctx, prefix+key).Err()
}

func (r *RedisCache) SetMonthExpense(key string, value int, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisCache) GetMonthExpense(key string) (int, error) {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	monthexpense, _ := strconv.Atoi(data)

	return monthexpense, nil
}

func (r *RedisCache) DeleteMonthExpense(key string) error {
	return r.client.Del(ctx, key).Err()
}
