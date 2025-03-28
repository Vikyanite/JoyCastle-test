package service

import (
	"example.com/m/v2/milirank/internal/model"
	"example.com/m/v2/milirank/internal/repository"
	"fmt"
	"time"

	"golang.org/x/time/rate"
)

type rankServiceImpl struct {
	redisRepo repository.RedisRepo
	fallback  repository.MysqlRepo
	limiter   *rate.Limiter
}

func NewRankService(redisRepo repository.RedisRepo, mysqlRepo repository.RankRepo) RankService {
	svc := &rankServiceImpl{
		redisRepo: redisRepo,
		fallback:  mysqlRepo,
		limiter:   rate.NewLimiter(rate.Every(100*time.Millisecond), 100),
	}
	return svc
}

func (s *rankServiceImpl) UpdateScore(userID string, score int64, timestamp int64) error {
	if !s.redisRepo.IsAlive() {
		if !s.limiter.Allow() {
			return fmt.Errorf("系统繁忙，请稍后再试")
		}

		return s.fallback.SetScore(userID, score, timestamp)
	}

	// 正常写 Redis
	return s.redisRepo.SetScore(userID, score, timestamp)
}

func (s *rankServiceImpl) GetTopN(n int64) ([]model.RankItem, error) {
	if !s.redisRepo.IsAlive() {
		if !s.limiter.Allow() {
			return nil, fmt.Errorf("系统繁忙，请稍后再试")
		}

		return s.fallback.GetTopN(n)
	}

	return s.redisRepo.GetTopN(n)
}

func (s *rankServiceImpl) GetUserRank(userID string) (*model.RankItem, error) {
	if !s.redisRepo.IsAlive() {
		if !s.limiter.Allow() {
			return nil, fmt.Errorf("系统繁忙，请稍后再试")
		}

		return s.fallback.GetUserRank(userID)
	}

	return s.redisRepo.GetUserRank(userID)
}

func (s *rankServiceImpl) GetUserRankRange(userID string, n int64) ([]model.RankItem, error) {
	if !s.redisRepo.IsAlive() {
		if !s.limiter.Allow() {
			return nil, fmt.Errorf("系统繁忙，请稍后再试")
		}

		return s.fallback.GetNearby(userID, n)
	}

	return s.redisRepo.GetNearby(userID, n)
}
