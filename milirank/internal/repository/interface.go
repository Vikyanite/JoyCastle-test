package repository

import "example.com/m/v2/milirank/internal/model"

type RedisRepo interface {
	RankRepo
	IsAlive() bool
}

type MysqlRepo interface {
	RankRepo
}

type RankRepo interface {
	SetScore(userID string, score int64, timestamp int64) error
	GetTopN(n int64) ([]model.RankItem, error)
	GetUserRank(userID string) (*model.RankItem, error)
	GetNearby(userID string, Range int64) ([]model.RankItem, error)
}
