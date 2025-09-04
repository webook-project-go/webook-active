package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/webook-project-go/webook-active/domain"
	"github.com/webook-project-go/webook-pkgs/logger"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client interface {
	SetActive(ctx context.Context, users []domain.User) error
	GetLastActiveAt(ctx context.Context, uid int64) (int64, error)
	JudgeActive(ctx context.Context, uid int64) (bool, error)
	GetActiveUsers(ctx context.Context, since, limit int64) ([]int64, error)
	ActiveFilters(ctx context.Context, uids []int64) ([]int64, error)
}

type client struct {
	cmd    redis.Cmdable
	client *cron.Cron
	l      logger.Logger
}

func New(cmd redis.Cmdable, l logger.Logger) Client {
	c := &client{
		cmd: cmd,
		l:   l,
	}
	err := c.StartCleaner(context.Background(), "0 0 3 * * *")
	if err != nil {
		panic(err)
	}
	return c
}
func (c *client) zsetKey() string {
	return "active:user:ls"
}
func (c *client) toDayKey() string {
	t := time.Now()
	return fmt.Sprintf("active:%04d-%02d-%02d", t.Year(), t.Month(), t.Day())

}
func (c *client) ActiveFilters(ctx context.Context, uids []int64) ([]int64, error) {
	key := c.toDayKey()
	pipe := c.cmd.Pipeline()
	cmds := make([]*redis.IntCmd, len(uids))

	for i, uid := range uids {
		cmds[i] = pipe.GetBit(ctx, key, uid)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	active := make([]int64, 0, len(uids))
	for i, cmd := range cmds {
		val, err := cmd.Result()
		if err != nil {
			return nil, err
		}
		if val == 1 {
			active = append(active, uids[i])
		}
	}

	return active, nil
}
func (c *client) SetActive(ctx context.Context, users []domain.User) error {
	day := c.toDayKey()
	zsetKey := c.zsetKey()

	tx := c.cmd.TxPipeline()
	z := make([]redis.Z, 0, len(users))
	for _, u := range users {
		tx.SetBit(ctx, day, u.UID, 1)
		z = append(z, redis.Z{
			Score:  float64(u.LastActive),
			Member: u.UID,
		})
	}
	tx.ZAdd(ctx, zsetKey, z...)
	_, err := tx.Exec(ctx)
	return err
}

func (c *client) GetLastActiveAt(ctx context.Context, uid int64) (int64, error) {
	score, err := c.cmd.ZScore(ctx, c.zsetKey(), strconv.FormatInt(uid, 10)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, err
	}
	return int64(score), nil
}

func (c *client) JudgeActive(ctx context.Context, uid int64) (bool, error) {
	val, err := c.cmd.GetBit(ctx, c.toDayKey(), uid).Result()
	if err != nil {
		return false, err
	}
	return val == 1, nil
}

func (c *client) GetActiveUsers(ctx context.Context, since int64, limit int64) ([]int64, error) {
	members, err := c.cmd.ZRangeByScore(ctx, c.zsetKey(), &redis.ZRangeBy{
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
	return c.cmd.ZRemRangeByScore(ctx, c.zsetKey(), "0", strconv.FormatInt(expireBefore, 10)).Err()
}

func (c *client) StartCleaner(ctx context.Context, spec string) error {
	if c.client == nil {
		c.client = cron.New(cron.WithSeconds())
	}

	_, err := c.client.AddFunc(spec, func() {
		if err := c.CleanExpired(ctx); err != nil {
			c.l.Error("CleanExpired error:", logger.Error(err))
		} else {
			c.l.Info("CleanExpired success at", logger.String("time:", time.Now().String()))
		}
	})

	if err != nil {
		return err
	}

	c.client.Start()
	return nil
}

func (c *client) StopCleaner() {
	if c.client != nil {
		c.client.Stop()
	}
}
