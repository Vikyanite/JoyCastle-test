package repository

import (
	"context"
	"errors"
	"example.com/m/v2/milirank/internal/model"
	"example.com/m/v2/redisx"
	"fmt"
	"github.com/redis/go-redis/v9"
)

var (
	RedisKey = "rank:million"
	ctx      = context.Background()
)

type redisRankDao struct {
	client *redisx.Client
}

func NewRankDao(client *redisx.Client) RedisRepo {
	return &redisRankDao{client: client}
}

func (r *redisRankDao) SetScore(playerID string, score int64, timestamp int64) error {
	return r.client.ZAdd(ctx, RedisKey, redis.Z{
		Score:  float64(score) + 1.0/float64(timestamp),
		Member: playerID,
	}).Err()
}

func (r *redisRankDao) IsAlive() bool {
	return r.client.IsAlive()
}

func (r *redisRankDao) GetTopN(N int64) ([]model.RankItem, error) {
	res, err := r.client.ZRevRangeWithScores(ctx, RedisKey, 0, N-1).Result()
	if err != nil {
		return nil, err
	}
	var items []model.RankItem
	for i, z := range res {
		items = append(items, model.RankItem{
			UserID: z.Member.(string),
			Score:  int64(z.Score),
			Rank:   int64(i + 1),
		})
	}
	return items, nil
}

func (r *redisRankDao) GetUserRank(playerID string) (*model.RankItem, error) {
	rank, err := r.client.ZRevRank(ctx, RedisKey, playerID).Result()
	if errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("player not found")
	}
	if err != nil {
		return nil, err
	}
	score, err := r.client.ZScore(ctx, RedisKey, playerID).Result()
	if err != nil {
		return nil, err
	}
	return &model.RankItem{
		UserID: playerID,
		Score:  int64(score),
		Rank:   rank + 1,
	}, nil
}

func (r *redisRankDao) GetNearby(playerID string, Range int64) ([]model.RankItem, error) {
	rank, err := r.client.ZRevRank(ctx, RedisKey, playerID).Result()
	if err != nil {
		return nil, err
	}

	start := int64(0)
	if rank > Range {
		start = rank - Range
	}
	end := rank + Range

	res, err := r.client.ZRevRangeWithScores(ctx, RedisKey, start, end).Result()
	if err != nil {
		return nil, err
	}

	var items []model.RankItem
	for i, z := range res {
		items = append(items, model.RankItem{
			UserID: z.Member.(string),
			Score:  int64(z.Score),
			Rank:   start + int64(i) + 1,
		})
	}
	return items, nil
}
