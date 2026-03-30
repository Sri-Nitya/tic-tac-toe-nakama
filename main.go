package main

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/heroiclabs/nakama-common/runtime"
)

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	logger.Info("NITYA TEST: MODULE LOADED")

	err := initializer.RegisterMatch(matchLabel, func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
		return &MatchHandler{}, nil
	})

	if err != nil {
		logger.Error("NITYA TEST: ERROR REGISTERING MATCH: %v", err)
		return err
	}

	err = initializer.RegisterRpc("create_match", func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		matchId, err := nk.MatchCreate(ctx, matchLabel, nil)
		if err != nil {
			logger.Error("NITYA TEST: ERROR CREATING MATCH: %v", err)
			return "", err
		}

		response := map[string]string{"matchId": matchId}

		responseBytes, _ := json.Marshal(response)
		
		return string(responseBytes), nil
	})

	if err != nil {
		logger.Error("NITYA TEST: ERROR REGISTERING RPC: %v", err)
		return err
	}

	err = initializer.RegisterRpc("quick_match", func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		minSize := 0
		maxSize := 2
		matches, err := nk.MatchList(ctx, 10, true, matchLabel, &minSize, &maxSize, "")
		if err != nil {
			logger.Error("ERROR LISTING MATCHES: %v", err)
			return "", err
		}

		for _, match := range matches {
			if match.Size < 2 {
				response := map[string]string{"matchId": match.MatchId}
				responseBytes, _ := json.Marshal(response)
				return string(responseBytes), nil
			}
		}
		matchId, err := nk.MatchCreate(ctx, matchLabel, nil)
		if err != nil {
			logger.Error("ERROR CREATING MATCH: %v", err)
			return "", err
		}
		response := map[string]string{"matchId": matchId}
		responseBytes, _ := json.Marshal(response)
		return string(responseBytes), nil
	})
	return nil
}
