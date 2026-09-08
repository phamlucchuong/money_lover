package auth

import (
	"context"
	"errors"

	"chuongpl/quan-ly-chi-tieu/internal/platform/cache"
)

var ErrBlacklistCheck = errors.New("blacklist check failed")

func blacklistKey(jti string) string {
	return "blacklist:" + jti
}

func NewBlacklistChecker(redisCache *cache.RedisCache) func(ctx context.Context, jti string) (bool, error) {
	return func(ctx context.Context, jti string) (bool, error) {
		exists, err := redisCache.Exists(ctx, blacklistKey(jti))
		if err != nil {
			return false, ErrBlacklistCheck
		}
		return exists, nil
	}
}
