package repository

import (
	"database/sql"
	"example.com/m/v2/milirank/internal/model"
	"time"
)

type mysqlRankDao struct {
	db *sql.DB
}

func NewRankDaoFallback(db *sql.DB) MysqlRepo {
	return &mysqlRankDao{db: db}
}

func (m *mysqlRankDao) SetScore(playerID string, score int64, timestamp int64) error {
	_, err := m.db.Exec(`
		INSERT INTO player_score (player_id, score, updated_at)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE score = VALUES(score), updated_at = VALUES(updated_at)
	`, playerID, score, time.Unix(timestamp, 0))
	return err
}

func (m *mysqlRankDao) GetTopN(N int64) ([]model.RankItem, error) {
	rows, err := m.db.Query(`
		SELECT player_id, score FROM player_score ORDER BY score DESC LIMIT ?
	`, N)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.RankItem
	var i int64 = 1
	for rows.Next() {
		var userID string
		var score int64
		if err := rows.Scan(&userID, &score); err != nil {
			return nil, err
		}
		items = append(items, model.RankItem{
			UserID: userID,
			Score:  score,
			Rank:   i,
		})
		i++
	}
	return items, nil
}

func (m *mysqlRankDao) GetUserRank(playerID string) (*model.RankItem, error) {
	var item model.RankItem
	err := m.db.QueryRow(`
		SELECT player_id, score,
			(SELECT COUNT(*) + 1 FROM player_score WHERE score > ps.score) as rank
		FROM player_score ps
		WHERE player_id = ?
	`, playerID).Scan(&item.UserID, &item.Score, &item.Rank)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *mysqlRankDao) GetNearby(playerID string, Range int64) ([]model.RankItem, error) {
	rankItem, err := m.GetUserRank(playerID)
	if err != nil {
		return nil, err
	}
	start := rankItem.Rank - Range
	if start < 1 {
		start = 1
	}

	rows, err := m.db.Query(`
		SELECT player_id, score, rk FROM (
			SELECT player_id, score, RANK() OVER (ORDER BY score DESC) as rk
			FROM player_score
		) t WHERE rk BETWEEN ? AND ?
	`, start, rankItem.Rank+Range)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.RankItem
	for rows.Next() {
		var userID string
		var score int64
		var iRank int64
		if err := rows.Scan(&userID, &score, &iRank); err != nil {
			return nil, err
		}
		items = append(items, model.RankItem{
			UserID: userID,
			Score:  score,
			Rank:   iRank,
		})
	}
	return items, nil
}
