package service

import "example.com/m/v2/milirank/internal/model"

type RankService interface {
	UpdateScore(userID string, score int64, timestamp int64) error
	GetUserRank(userID string) (*model.RankItem, error)
	GetTopN(N int64) ([]model.RankItem, error)
	GetUserRankRange(userID string, Range int64) ([]model.RankItem, error)
}
