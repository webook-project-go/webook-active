package redis

import (
	"context"
	"errors"
	"github.com/webook-project-go/webook-active/domain"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client interface {
	SetActive(ctx context.Context, users []domain.User) error
	GetLastActiveAt(ctx context.Context, uid int64) (int64, error)
	JudgeActive(ctx context.Context, uid int64) (bool, error)
	GetActiveUsers(ctx context.Context, since, limit int64) ([]int64, error)
}

type client struct {
	cmd redis.Cmdable
	key string
}

func New(cmd redis.Cmdable) Client {
	return &client{
		cmd: cmd,
		key: "user_active_ts",
	}
}

func (c *client) SetActive(ctx context.Context, users []domain.User) error {
	z := make([]redis.Z, 0, len(users))
	for i := 0; i < len(users); i++ {
		z = append(z, redis.Z{
			Score:  float64(users[i].LastActive),
			Member: users[i].UID,
		})
	}
	return c.cmd.ZAdd(ctx, c.key, z...).Err()
}

func (c *client) GetLastActiveAt(ctx context.Context, uid int64) (int64, error) {
	score, err := c.cmd.ZScore(ctx, c.key, strconv.FormatInt(uid, 10)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, err
	}
	return int64(score), nil
}

func (c *client) JudgeActive(ctx context.Context, uid int64) (bool, error) {
	lastActive, err := c.GetLastActiveAt(ctx, uid)
	if err != nil {
		return false, err
	}
	if lastActive == 0 {
		return false, nil
	}
	return time.Now().Unix()-lastActive <= 3*24*3600, nil
}

func (c *client) GetActiveUsers(ctx context.Context, since int64, limit int64) ([]int64, error) {
	members, err := c.cmd.ZRangeByScore(ctx, c.key, &redis.ZRangeBy{
		Min:    strconv.FormatInt(since, 10),
		Max:    "+inf",
		Offset: 0,
		Count:  int64(limit),
	}).Result()
	if err != nil {
		return nil, err
	}

	var uids []int64
	for _, m := range members {
		uid, err := strconv.ParseInt(m, 10, 64)
		if err != nil {
			continue
		}
		uids = append(uids, uid)
	}
	return uids, nil
}

func (c *client) CleanExpired(ctx context.Context) error {
	expireBefore := time.Now().Add(-72 * time.Hour).Unix()
	return c.cmd.ZRemRangeByScore(ctx, c.key, "0", strconv.FormatInt(expireBefore, 10)).Err()
}
