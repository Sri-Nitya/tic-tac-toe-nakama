package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"

	"github.com/heroiclabs/nakama-common/runtime"
)

type PlayerStats struct {
	UserID    string `json:"userId"`
	Wins      int    `json:"wins"`
	Losses    int    `json:"losses"`
	Draws     int    `json:"draws"`
	WinStreak int    `json:"winStreak"`
}

const statsCollection = "player_stats"
const statsKey = "summary"

func getPlayerStats(ctx context.Context, nk runtime.NakamaModule, userID string) (*PlayerStats, error) {
	objects, err := nk.StorageRead(ctx, []*runtime.StorageRead{
		{
			Collection: statsCollection,
			Key:        statsKey,
			UserID:     userID,
		},
	})
	if err != nil {
		return nil, err
	}

	if len(objects) == 0 {
		return &PlayerStats{
			UserID:    userID,
			Wins:      0,
			Losses:    0,
			Draws:     0,
			WinStreak: 0,
		}, nil
	}

	var stats PlayerStats
	if err := json.Unmarshal([]byte(objects[0].Value), &stats); err != nil {
		return nil, err
	}

	if stats.UserID == "" {
		stats.UserID = userID
	}

	return &stats, nil
}

func savePlayerStats(ctx context.Context, nk runtime.NakamaModule, stats *PlayerStats) error {
	value, err := json.Marshal(stats)
	if err != nil {
		return err
	}

	_, err = nk.StorageWrite(ctx, []*runtime.StorageWrite{
		{
			Collection:      statsCollection,
			Key:             statsKey,
			UserID:          stats.UserID,
			Value:           string(value),
			PermissionRead:  2,
			PermissionWrite: 0,
		},
	})

	return err
}

func getPlayerIDsFromState(s *MatchState) []string {
	playerIDs := make([]string, 0, len(s.players))
	for userID := range s.players {
		playerIDs = append(playerIDs, userID)
	}
	return playerIDs
}

func updatePlayerStatsForResult(ctx context.Context, nk runtime.NakamaModule, winnerID string, playerIDs []string, isDraw bool) error {
	for _, userID := range playerIDs {
		stats, err := getPlayerStats(ctx, nk, userID)
		if err != nil {
			return err
		}

		if isDraw {
			stats.Draws++
			stats.WinStreak = 0
		} else if userID == winnerID {
			stats.Wins++
			stats.WinStreak++
		} else {
			stats.Losses++
			stats.WinStreak = 0
		}

		if err := savePlayerStats(ctx, nk, stats); err != nil {
			return err
		}
	}

	return nil
}

func getLeaderboardRpc(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	limit := 100
	var cursor string

	type LeaderboardEntry struct {
		UserID    string `json:"userId"`
		Wins      int    `json:"wins"`
		Losses    int    `json:"losses"`
		Draws     int    `json:"draws"`
		WinStreak int    `json:"winStreak"`
	}

	entries := make([]LeaderboardEntry, 0)

	for {
		objects, nextCursor, err := nk.StorageList(ctx, "", "", statsCollection, limit, cursor)
		if err != nil {
			logger.Error("failed to list stats: %v", err)
			return "", err
		}

		for _, obj := range objects {
			var stats PlayerStats
			if err := json.Unmarshal([]byte(obj.Value), &stats); err != nil {
				logger.Error("failed to parse stats object: %v", err)
				continue
			}

			if stats.UserID == "" {
				stats.UserID = obj.UserId
			}

			entries = append(entries, LeaderboardEntry{
				UserID:    stats.UserID,
				Wins:      stats.Wins,
				Losses:    stats.Losses,
				Draws:     stats.Draws,
				WinStreak: stats.WinStreak,
			})
		}

		if nextCursor == "" {
			break
		}
		cursor = nextCursor
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Wins != entries[j].Wins {
			return entries[i].Wins > entries[j].Wins
		}
		if entries[i].WinStreak != entries[j].WinStreak {
			return entries[i].WinStreak > entries[j].WinStreak
		}
		if entries[i].Losses != entries[j].Losses {
			return entries[i].Losses < entries[j].Losses
		}
		return entries[i].UserID < entries[j].UserID
	})

	result, err := json.Marshal(map[string]interface{}{
		"leaders": entries,
	})
	if err != nil {
		return "", err
	}

	return string(result), nil
}