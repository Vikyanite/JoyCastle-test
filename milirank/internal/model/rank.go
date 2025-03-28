package model

type RankItem struct {
	UserID string `json:"user_id"`
	Score  int64  `json:"score"`
	Rank   int64  `json:"rank"`
}
